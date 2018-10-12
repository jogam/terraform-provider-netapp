# test implementation of key store

from concurrent import futures
import sys
import time

import grpc

import keystore_pb2
import keystore_pb2_grpc

from grpc_health.v1.health import HealthServicer
from grpc_health.v1 import health_pb2, health_pb2_grpc


def msg(text):
    print(text)
    sys.stdout.flush()

class KeyStoreServicer(keystore_pb2_grpc.KeyStoreServicer):
    """Implementation of KV service."""

    def Get(self, request, context):
        filename = "kv_"+request.key
        msg("GET request: " + request.key)
        with open(filename, 'r+b') as f:
            result = keystore_pb2.GetResponse()
            result.value = f.read()
            return result

    def Put(self, request, context):
        msg("PUT request: " + request.key + " = " + str(request.value))
        filename = "kv_"+request.key
        value = "{0}\n\nWritten from plugin-python".format(request.value)
        with open(filename, 'w') as f:
            f.write(value)

        return keystore_pb2.Empty()


def serve():
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
    server.add_insecure_port('127.0.0.1:1234')
    server.start()

    # Output information
    msg("1|1|tcp|127.0.0.1:1234|grpc")

    try:
        while True:
            time.sleep(60 * 60 * 24)
    except KeyboardInterrupt:
        server.stop(0)

        msg("exiting python keystore server")

    # make a change

if __name__ == '__main__':
    serve()
