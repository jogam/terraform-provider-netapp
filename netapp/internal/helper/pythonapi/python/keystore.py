# test implementation of key store

import logging
import os.path

from concurrent import futures
import sys
import time

import grpc

import keystore_pb2
import keystore_pb2_grpc

from grpc_health.v1.health import HealthServicer
from grpc_health.v1 import health_pb2, health_pb2_grpc


logging.basicConfig(filename='python_api.log',level=logging.DEBUG)
SHUTDOWN_FILE = "./shut_api"

def msg(text):
    print(text)
    sys.stdout.flush()

class KeyStoreServicer(keystore_pb2_grpc.KeyStoreServicer):
    """Implementation of KV service."""

    def Get(self, request, context):
        filename = "kv_"+request.key
        logging.debug("GET request: " + request.key)
        with open(filename, 'r') as f:
            result = keystore_pb2.GetResponse()
            result.value = f.read()
            return result

    def Put(self, request, context):
        logging.debug("PUT request: " + request.key + " = " + request.value)
        filename = "kv_"+request.key
        with open(filename, 'w') as f:
            f.write(request.value)

        return keystore_pb2.Empty()


def serve(host='127.0.0.1', port='1234'):
    # We need to build a health service to work with go-plugin
    health = HealthServicer()
    health.set(
        "plugin",
        health_pb2.HealthCheckResponse.ServingStatus.Value('SERVING'))

    # Start the server.
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    keystore_pb2_grpc.add_KeyStoreServicer_to_server(
                            KeyStoreServicer(), server)
    health_pb2_grpc.add_HealthServicer_to_server(health, server)
    server.add_insecure_port(host + ':' + port)
    server.start()

    # Output information
    msg("1|1|tcp|" + host + ":" + port + "|grpc")

    try:
        while True:
            if os.path.isfile(SHUTDOWN_FILE):
                break

            time.sleep(5)

    except KeyboardInterrupt:
        pass

    logging.debug("left while loop, issuing server.stop()")        
    server.stop(0)
    logging.debug("exiting python keystore server")

    # make a change

if __name__ == '__main__':
    serve()
