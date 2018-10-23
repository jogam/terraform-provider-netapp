# test implementation of key store

import logging
import os
import sys

import sys

# append the NetApp SDK python library to path
sys.path.append(
    os.path.join(
        os.environ.get('NETAPP_MSDK_ROOT_PATH', 'FAULT'),
        'lib', 'python', 'NetApp'))

from multiprocessing import Event
import socket
from contextlib import closing


from concurrent import futures
import sys
import time

from util import (
    RegistryServer, 
    ClientStatus, 
    ApiStatus,
    CallCounter
)

import grpc

import grpcapi_pb2
import grpcapi_pb2_grpc

from grpc_health.v1.health import HealthServicer
from grpc_health.v1 import health_pb2, health_pb2_grpc


logging.basicConfig(
    filename='python_api.log',
    level=logging.DEBUG,
    format=(
        '[%(asctime)s %(levelname)s][' + str(os.getpid()) 
        + '] @{%(lineno)d} - %(message)s')
)

LOCALHOST = "172.0.0.1"
CHECK_TIMEOUT = 0.8             # check if api is being used every 800ms
RUNNING_FILE = "./API_UP"       # file indicating to outside that API grpc server is up

class NetAppApiServicer(grpcapi_pb2_grpc.GRPCNetAppApiServicer):
    """Implementation of NetAppApiServicer."""

    def __init__(self, registry, counter, *args, **kwargs):
        super(NetAppApiServicer, self).__init__(*args, **kwargs)
        self.registry = registry
        self.counter = counter
        logging.debug("servicer initialized with call counter")

    def Call(self, request, context):
        logging.debug("Call request: %s", request.cmd)
        self.counter.increment()
        resp = grpcapi_pb2.CallResponse()
        resp.success = True
        resp.errmsg = ''
        resp.data = b'something'
        return resp

    def Shutdown(self, request, context):
        logging.debug("SD request for client: %s", request.clientid)
        success = self.registry.set_client_status(
            request.clientid,
            ClientStatus.shutdown)
        resp = grpcapi_pb2.ShutdownResponse()
        resp.result = success
        return resp

def notify_grpc(client_id, registry):
    # send stdout msg for go-plugin to understand...
    grpc_msg = registry.get_grpc_msg()
    logging.info("client [%s] gRPC message: %s", client_id, grpc_msg)
    print(grpc_msg)
    sys.stdout.flush()

def register_client(registry, client_id):
    if not registry.register_client(client_id, ClientStatus.running):
        logging.error(
            'netapp API could not register client: %s', client_id)
        return False

    return True

def unregister_client(registry, client_id):
    status = registry.unregister_client(client_id)
    if ClientStatus.is_error(status):
        logging.error(
            'registry unregister client status error: %s', status.value)
        return False

    return True

def serve(client_id, host='127.0.0.1', port='1234'):

    # create client registry server
    serv_event = Event()
    reg_server = RegistryServer(serv_event)
    reg_server.start()

    while not serv_event.is_set():
        serv_event.wait(1)
    serv_event.clear()

    # create client registry
    registry = RegistryServer.GET_CLIENT_REGISTRY()

    # create a call counter for call to API
    call_counter = CallCounter(initval=1)

    # create the servicer with this semaphore
    servicer = NetAppApiServicer(registry, call_counter)

    # generate gRPC connection message and store in registry
    grpc_msg = "1|1|tcp|" + host + ":" + port + "|grpc"
    if not registry.set_grpc_msg(grpc_msg):
        logging.error('could not set gRPC message in registry')

        registry.shutdown()

        while not serv_event.is_set():
            serv_event.wait(1)

        reg_server.terminate()
        reg_server.join()
        sys.exit(1)

    # register client
    if not register_client(registry, client_id):
        logging.error('register server start client in registry')

        registry.shutdown()

        while not serv_event.is_set():
            serv_event.wait(1)

        reg_server.terminate()
        reg_server.join()
        sys.exit(1)

    # We need to build a health service to work with go-plugin
    health = HealthServicer()
    health.set(
        "plugin",
        health_pb2.HealthCheckResponse.ServingStatus.Value('SERVING'))

    # Start the server.
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    grpcapi_pb2_grpc.add_GRPCNetAppApiServicer_to_server(servicer, server)
    health_pb2_grpc.add_HealthServicer_to_server(health, server)
    server.add_insecure_port(host + ':' + port)
    server.start()

    logging.debug('started GRPC server @ %s:%s', host, port)

    # let GRPC know we are here and good...
    notify_grpc(client_id, registry)

    # set registry status to running
    if not registry.set_api_status(ApiStatus.running):
        logging.error('could not set api status in registry')

        server.stop(0)
        registry.shutdown()

        while not serv_event.is_set():
            serv_event.wait(1)

        reg_server.terminate()
        reg_server.join()
        sys.exit(1)

    # create running status file
    # source: https://stackoverflow.com/a/12654798
    with open(RUNNING_FILE, 'a'):
        os.utime(RUNNING_FILE, None)

    logging.debug('running file created')

    try:
        while call_counter.value() > 0:
            # get the number of calls
            # NOTE: first run will be 1 to not step out
            call_cnt = call_counter.value()
            # reset count calls (NOTE: to 0!)
            call_counter.reset()
            
            logging.debug("was needed %d times, looping...", call_cnt)

            # wait for CHECK_TIMEOUT to receive calls
            time.sleep(CHECK_TIMEOUT)

    except KeyboardInterrupt:
        pass

    logging.debug("left while loop, issuing server.stop()")   
    unregister_client(registry, client_id)
    server.stop(0)
    os.remove(RUNNING_FILE)     # making sure we are not running
    registry.shutdown()

    # sleep until server indicates shutdown
    while not serv_event.is_set():
        serv_event.wait(1)

    reg_server.terminate()
    reg_server.join()
    logging.debug("exiting netapp API serve()")


if __name__ == '__main__':

    # get length of provided arguments after script path
    arg_cnt = len(sys.argv) - 1

    if arg_cnt != 2:
        logging.error('netapp API must be called as: api.py PORT CLIENTID!')
        sys.exit(1)

    parent_pid = os.getppid()
    api_pid = os.getpid()

    logging.info('netapp API starting (PPID, PID): [%d, %d]', parent_pid, api_pid)

    # retrieve parameters
    api_port = sys.argv[1]
    client_id = sys.argv[2]

    if os.path.exists(RUNNING_FILE):
        # alreay running, lets do something ugly
        logging.warn('netapp API already running, doing file waiting...')

        # get client registry and verify its still in running status
        registry = RegistryServer.GET_CLIENT_REGISTRY()
        api_status = registry.get_api_status()
        if api_status != ApiStatus.running:
            logging.error('netapp API not running per client registry')
            sys.exit(1)

        # register client with registry
        if not register_client(registry, client_id):
            sys.exit(1)

        # let GRPC know we are here and good...
        notify_grpc(client_id, registry)

        # wait until client status is shutdown
        while True:
            status = registry.get_client_status(client_id)
            if ClientStatus.is_error(status):
                logging.error(
                    'registry get client status error: %s', status.value)
                break
            
            if status == ClientStatus.shutdown:
                break

            # wait for new calls
            time.sleep(CHECK_TIMEOUT)

        # unregister client from registry
        unregister_client(registry, client_id)
    else:
        logging.warn('netapp API not running, start server')
        serve(client_id, port=api_port)

    logging.info(
        'netapp API [%s] exiting (PPID, PID): [%d, %d]', 
        client_id, parent_pid, api_pid)
