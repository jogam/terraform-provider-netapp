# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: grpcapi.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor.FileDescriptor(
  name='grpcapi.proto',
  package='grpcapi',
  syntax='proto3',
  serialized_options=None,
  serialized_pb=_b('\n\rgrpcapi.proto\x12\x07grpcapi\"(\n\x0b\x43\x61llRequest\x12\x0b\n\x03\x63md\x18\x01 \x01(\t\x12\x0c\n\x04\x64\x61ta\x18\x02 \x01(\x0c\"=\n\x0c\x43\x61llResponse\x12\x0f\n\x07success\x18\x01 \x01(\x08\x12\x0e\n\x06\x65rrmsg\x18\x02 \x01(\t\x12\x0c\n\x04\x64\x61ta\x18\x03 \x01(\x0c\"#\n\x0fShutdownRequest\x12\x10\n\x08\x63lientid\x18\x01 \x01(\t\"\"\n\x10ShutdownResponse\x12\x0e\n\x06result\x18\x01 \x01(\x08\x32\x85\x01\n\rGRPCNetAppApi\x12\x33\n\x04\x43\x61ll\x12\x14.grpcapi.CallRequest\x1a\x15.grpcapi.CallResponse\x12?\n\x08Shutdown\x12\x18.grpcapi.ShutdownRequest\x1a\x19.grpcapi.ShutdownResponseb\x06proto3')
)




_CALLREQUEST = _descriptor.Descriptor(
  name='CallRequest',
  full_name='grpcapi.CallRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='cmd', full_name='grpcapi.CallRequest.cmd', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='data', full_name='grpcapi.CallRequest.data', index=1,
      number=2, type=12, cpp_type=9, label=1,
      has_default_value=False, default_value=_b(""),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=26,
  serialized_end=66,
)


_CALLRESPONSE = _descriptor.Descriptor(
  name='CallResponse',
  full_name='grpcapi.CallResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='success', full_name='grpcapi.CallResponse.success', index=0,
      number=1, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='errmsg', full_name='grpcapi.CallResponse.errmsg', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='data', full_name='grpcapi.CallResponse.data', index=2,
      number=3, type=12, cpp_type=9, label=1,
      has_default_value=False, default_value=_b(""),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=68,
  serialized_end=129,
)


_SHUTDOWNREQUEST = _descriptor.Descriptor(
  name='ShutdownRequest',
  full_name='grpcapi.ShutdownRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='clientid', full_name='grpcapi.ShutdownRequest.clientid', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=131,
  serialized_end=166,
)


_SHUTDOWNRESPONSE = _descriptor.Descriptor(
  name='ShutdownResponse',
  full_name='grpcapi.ShutdownResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='result', full_name='grpcapi.ShutdownResponse.result', index=0,
      number=1, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=168,
  serialized_end=202,
)

DESCRIPTOR.message_types_by_name['CallRequest'] = _CALLREQUEST
DESCRIPTOR.message_types_by_name['CallResponse'] = _CALLRESPONSE
DESCRIPTOR.message_types_by_name['ShutdownRequest'] = _SHUTDOWNREQUEST
DESCRIPTOR.message_types_by_name['ShutdownResponse'] = _SHUTDOWNRESPONSE
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

CallRequest = _reflection.GeneratedProtocolMessageType('CallRequest', (_message.Message,), dict(
  DESCRIPTOR = _CALLREQUEST,
  __module__ = 'grpcapi_pb2'
  # @@protoc_insertion_point(class_scope:grpcapi.CallRequest)
  ))
_sym_db.RegisterMessage(CallRequest)

CallResponse = _reflection.GeneratedProtocolMessageType('CallResponse', (_message.Message,), dict(
  DESCRIPTOR = _CALLRESPONSE,
  __module__ = 'grpcapi_pb2'
  # @@protoc_insertion_point(class_scope:grpcapi.CallResponse)
  ))
_sym_db.RegisterMessage(CallResponse)

ShutdownRequest = _reflection.GeneratedProtocolMessageType('ShutdownRequest', (_message.Message,), dict(
  DESCRIPTOR = _SHUTDOWNREQUEST,
  __module__ = 'grpcapi_pb2'
  # @@protoc_insertion_point(class_scope:grpcapi.ShutdownRequest)
  ))
_sym_db.RegisterMessage(ShutdownRequest)

ShutdownResponse = _reflection.GeneratedProtocolMessageType('ShutdownResponse', (_message.Message,), dict(
  DESCRIPTOR = _SHUTDOWNRESPONSE,
  __module__ = 'grpcapi_pb2'
  # @@protoc_insertion_point(class_scope:grpcapi.ShutdownResponse)
  ))
_sym_db.RegisterMessage(ShutdownResponse)



_GRPCNETAPPAPI = _descriptor.ServiceDescriptor(
  name='GRPCNetAppApi',
  full_name='grpcapi.GRPCNetAppApi',
  file=DESCRIPTOR,
  index=0,
  serialized_options=None,
  serialized_start=205,
  serialized_end=338,
  methods=[
  _descriptor.MethodDescriptor(
    name='Call',
    full_name='grpcapi.GRPCNetAppApi.Call',
    index=0,
    containing_service=None,
    input_type=_CALLREQUEST,
    output_type=_CALLRESPONSE,
    serialized_options=None,
  ),
  _descriptor.MethodDescriptor(
    name='Shutdown',
    full_name='grpcapi.GRPCNetAppApi.Shutdown',
    index=1,
    containing_service=None,
    input_type=_SHUTDOWNREQUEST,
    output_type=_SHUTDOWNRESPONSE,
    serialized_options=None,
  ),
])
_sym_db.RegisterServiceDescriptor(_GRPCNETAPPAPI)

DESCRIPTOR.services_by_name['GRPCNetAppApi'] = _GRPCNETAPPAPI

# @@protoc_insertion_point(module_scope)
