# test implementation of key store

import logging
import os
import sys

from multiprocessing import Value, Lock
import socket
from contextlib import closing


from concurrent import futures
import sys
import time

import grpc

import keystore_pb2
import keystore_pb2_grpc

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
CHECK_TIMEOUT = 0.5             # check if api is being used every 0.5 seconds
RUNNING_FILE = "./API_UP"       # file indicating to outside that API grpc server is up
SHUTDOWN_FILE = "./shut_api"    # file flagging to service to shutdown...


class CallCounter(object):
    '''
    multiprocessing/threading save counter
    inspired by: https://eli.thegreenplace.net/2012/01/04/shared-counter-with-pythons-multiprocessing
    '''

    def __init__(self, initval=0):
        self.val = Value('i', initval)
        self.lock = Lock()

    def increment(self):
        with self.lock:
            self.val.value += 1

    def value(self):
        with self.lock:
            return self.val.value

    def reset(self, initval=0):
        with self.lock:
            self.val.value = initval


class KeyStoreServicer(keystore_pb2_grpc.KeyStoreServicer):
    """Implementation of KV service."""

    def __init__(self, counter, *args, **kwargs):
        super(KeyStoreServicer, self).__init__(*args, **kwargs)
        self.counter = counter
        logging.debug("servicer initialized with call counter")

    def Get(self, request, context):
        filename = "kv_"+request.key
        logging.debug("GET request: " + request.key)
        self.counter.increment()
        with open(filename, 'r') as f:
            result = keystore_pb2.GetResponse()
            result.value = f.read()
            return result

    def Put(self, request, context):
        logging.debug("PUT request: " + request.key + " = " + request.value)
        filename = "kv_"+request.key
        self.counter.increment()
        with open(filename, 'w') as f:
            f.write(request.value)

        return keystore_pb2.Empty()

def api_up(host, port):
    connected = False
    apiport = int(port)
    with closing(socket.socket(socket.AF_INET, socket.SOCK_STREAM)) as sock:
        sock.settimeout(0.4)    # 400 ms timeout here
        errno = sock.connect_ex((host, apiport))
        logging.debug('sock.connect returned: %d', errno)
        connected = (errno == 0)

    return connected

def serve(host='127.0.0.1', port='1234'):

    # create a call counter for call to API
    call_counter = CallCounter(initval=1)

    # create the servicer with this semaphore
    servicer = KeyStoreServicer(call_counter)

    # We need to build a health service to work with go-plugin
    health = HealthServicer()
    health.set(
        "plugin",
        health_pb2.HealthCheckResponse.ServingStatus.Value('SERVING'))

    # Start the server.
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    keystore_pb2_grpc.add_KeyStoreServicer_to_server(servicer, server)
    health_pb2_grpc.add_HealthServicer_to_server(health, server)
    server.add_insecure_port(host + ':' + port)
    server.start()

    # send stdout msg for go-plugin to understand...
    grpc_msg = "1|1|tcp|" + host + ":" + port + "|grpc"
    logging.info("GRPC message: %s", grpc_msg)
    print(grpc_msg)
    sys.stdout.flush()

    # create running status file
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

            if os.path.isfile(SHUTDOWN_FILE):
                logging.debug('received shutdown file trigger')
                os.remove(SHUTDOWN_FILE)
                break

            # wait for CHECK_TIMEOUT to receive calls
            time.sleep(CHECK_TIMEOUT)

    except KeyboardInterrupt:
        pass

    logging.debug("left while loop, issuing server.stop()")        
    server.stop(0)
    os.remove(RUNNING_FILE)     # making sure we are not running
    logging.debug("exiting netapp API serve()")


if __name__ == '__main__':

    # get length of provided arguments after script path
    arg_cnt = len(sys.argv) - 1

    if arg_cnt != 1:
        logging.error('netapp API must be called as: api.py PORT!')
        sys.exit(1)

    parent_pid = os.getppid()
    api_pid = os.getpid()

    logging.info('netapp API starting (PPID, PID): [%d, %d]', parent_pid, api_pid)

    # retrieve parameters
    api_port = sys.argv[1]
    if api_up(LOCALHOST, api_port):
        # alreay running, lets do something ugly
        # TODO: use ReattachConfig from GRPC instead?
        logging.warn('netapp API already running, doing file waiting...')
        while os.path.exists(RUNNING_FILE):
            time.sleep(CHECK_TIMEOUT)
    else:
        logging.warn('netapp API not running, start server')
        serve(port=api_port)

    logging.info('netapp API exiting (PPID, PID): [%d, %d]', parent_pid, api_pid)
