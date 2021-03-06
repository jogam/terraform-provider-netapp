# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
import grpc

import grpcapi_pb2 as grpcapi__pb2


class GRPCNetAppApiStub(object):
  """the simple API definition
  """

  def __init__(self, channel):
    """Constructor.

    Args:
      channel: A grpc.Channel.
    """
    self.Call = channel.unary_unary(
        '/grpcapi.GRPCNetAppApi/Call',
        request_serializer=grpcapi__pb2.CallRequest.SerializeToString,
        response_deserializer=grpcapi__pb2.CallResponse.FromString,
        )
    self.Shutdown = channel.unary_unary(
        '/grpcapi.GRPCNetAppApi/Shutdown',
        request_serializer=grpcapi__pb2.ShutdownRequest.SerializeToString,
        response_deserializer=grpcapi__pb2.ShutdownResponse.FromString,
        )


class GRPCNetAppApiServicer(object):
  """the simple API definition
  """

  def Call(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')

  def Shutdown(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')


def add_GRPCNetAppApiServicer_to_server(servicer, server):
  rpc_method_handlers = {
      'Call': grpc.unary_unary_rpc_method_handler(
          servicer.Call,
          request_deserializer=grpcapi__pb2.CallRequest.FromString,
          response_serializer=grpcapi__pb2.CallResponse.SerializeToString,
      ),
      'Shutdown': grpc.unary_unary_rpc_method_handler(
          servicer.Shutdown,
          request_deserializer=grpcapi__pb2.ShutdownRequest.FromString,
          response_serializer=grpcapi__pb2.ShutdownResponse.SerializeToString,
      ),
  }
  generic_handler = grpc.method_handlers_generic_handler(
      'grpcapi.GRPCNetAppApi', rpc_method_handlers)
  server.add_generic_rpc_handlers((generic_handler,))
