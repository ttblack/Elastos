# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
import grpc

from . import common_pb2 as common__pb2


class CommonStub(object):
  """The common service definition.
  """

  def __init__(self, channel):
    """Constructor.

    Args:
      channel: A grpc.Channel.
    """
    self.GenerateAPIRequest = channel.unary_unary(
        '/common.Common/GenerateAPIRequest',
        request_serializer=common__pb2.Request.SerializeToString,
        response_deserializer=common__pb2.Response.FromString,
        )
    self.GetAPIKey = channel.unary_unary(
        '/common.Common/GetAPIKey',
        request_serializer=common__pb2.Request.SerializeToString,
        response_deserializer=common__pb2.Response.FromString,
        )


class CommonServicer(object):
  """The common service definition.
  """

  def GenerateAPIRequest(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')

  def GetAPIKey(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')


def add_CommonServicer_to_server(servicer, server):
  rpc_method_handlers = {
      'GenerateAPIRequest': grpc.unary_unary_rpc_method_handler(
          servicer.GenerateAPIRequest,
          request_deserializer=common__pb2.Request.FromString,
          response_serializer=common__pb2.Response.SerializeToString,
      ),
      'GetAPIKey': grpc.unary_unary_rpc_method_handler(
          servicer.GetAPIKey,
          request_deserializer=common__pb2.Request.FromString,
          response_serializer=common__pb2.Response.SerializeToString,
      ),
  }
  generic_handler = grpc.method_handlers_generic_handler(
      'common.Common', rpc_method_handlers)
  server.add_generic_rpc_handlers((generic_handler,))