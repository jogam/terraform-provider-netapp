# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
import grpc

import keystore_pb2 as keystore__pb2


class KeyStoreStub(object):
  """the key store service definition
  """

  def __init__(self, channel):
    """Constructor.

    Args:
      channel: A grpc.Channel.
    """
    self.Get = channel.unary_unary(
        '/keystore.KeyStore/Get',
        request_serializer=keystore__pb2.GetRequest.SerializeToString,
        response_deserializer=keystore__pb2.GetResponse.FromString,
        )
    self.Put = channel.unary_unary(
        '/keystore.KeyStore/Put',
        request_serializer=keystore__pb2.PutRequest.SerializeToString,
        response_deserializer=keystore__pb2.Empty.FromString,
        )


class KeyStoreServicer(object):
  """the key store service definition
  """

  def Get(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')

  def Put(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')


def add_KeyStoreServicer_to_server(servicer, server):
  rpc_method_handlers = {
      'Get': grpc.unary_unary_rpc_method_handler(
          servicer.Get,
          request_deserializer=keystore__pb2.GetRequest.FromString,
          response_serializer=keystore__pb2.GetResponse.SerializeToString,
      ),
      'Put': grpc.unary_unary_rpc_method_handler(
          servicer.Put,
          request_deserializer=keystore__pb2.PutRequest.FromString,
          response_serializer=keystore__pb2.Empty.SerializeToString,
      ),
  }
  generic_handler = grpc.method_handlers_generic_handler(
      'keystore.KeyStore', rpc_method_handlers)
  server.add_generic_rpc_handlers((generic_handler,))
