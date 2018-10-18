# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: keystore.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor.FileDescriptor(
  name='keystore.proto',
  package='keystore',
  syntax='proto3',
  serialized_options=None,
  serialized_pb=_b('\n\x0ekeystore.proto\x12\x08keystore\"\x19\n\nGetRequest\x12\x0b\n\x03key\x18\x01 \x01(\t\"\x1c\n\x0bGetResponse\x12\r\n\x05value\x18\x01 \x01(\t\"(\n\nPutRequest\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t\"\x07\n\x05\x45mpty2l\n\x08KeyStore\x12\x32\n\x03Get\x12\x14.keystore.GetRequest\x1a\x15.keystore.GetResponse\x12,\n\x03Put\x12\x14.keystore.PutRequest\x1a\x0f.keystore.Emptyb\x06proto3')
)




_GETREQUEST = _descriptor.Descriptor(
  name='GetRequest',
  full_name='keystore.GetRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='key', full_name='keystore.GetRequest.key', index=0,
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
  serialized_start=28,
  serialized_end=53,
)


_GETRESPONSE = _descriptor.Descriptor(
  name='GetResponse',
  full_name='keystore.GetResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='value', full_name='keystore.GetResponse.value', index=0,
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
  serialized_start=55,
  serialized_end=83,
)


_PUTREQUEST = _descriptor.Descriptor(
  name='PutRequest',
  full_name='keystore.PutRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='key', full_name='keystore.PutRequest.key', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='value', full_name='keystore.PutRequest.value', index=1,
      number=2, type=9, cpp_type=9, label=1,
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
  serialized_start=85,
  serialized_end=125,
)


_EMPTY = _descriptor.Descriptor(
  name='Empty',
  full_name='keystore.Empty',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
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
  serialized_start=127,
  serialized_end=134,
)

DESCRIPTOR.message_types_by_name['GetRequest'] = _GETREQUEST
DESCRIPTOR.message_types_by_name['GetResponse'] = _GETRESPONSE
DESCRIPTOR.message_types_by_name['PutRequest'] = _PUTREQUEST
DESCRIPTOR.message_types_by_name['Empty'] = _EMPTY
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

GetRequest = _reflection.GeneratedProtocolMessageType('GetRequest', (_message.Message,), dict(
  DESCRIPTOR = _GETREQUEST,
  __module__ = 'keystore_pb2'
  # @@protoc_insertion_point(class_scope:keystore.GetRequest)
  ))
_sym_db.RegisterMessage(GetRequest)

GetResponse = _reflection.GeneratedProtocolMessageType('GetResponse', (_message.Message,), dict(
  DESCRIPTOR = _GETRESPONSE,
  __module__ = 'keystore_pb2'
  # @@protoc_insertion_point(class_scope:keystore.GetResponse)
  ))
_sym_db.RegisterMessage(GetResponse)

PutRequest = _reflection.GeneratedProtocolMessageType('PutRequest', (_message.Message,), dict(
  DESCRIPTOR = _PUTREQUEST,
  __module__ = 'keystore_pb2'
  # @@protoc_insertion_point(class_scope:keystore.PutRequest)
  ))
_sym_db.RegisterMessage(PutRequest)

Empty = _reflection.GeneratedProtocolMessageType('Empty', (_message.Message,), dict(
  DESCRIPTOR = _EMPTY,
  __module__ = 'keystore_pb2'
  # @@protoc_insertion_point(class_scope:keystore.Empty)
  ))
_sym_db.RegisterMessage(Empty)



_KEYSTORE = _descriptor.ServiceDescriptor(
  name='KeyStore',
  full_name='keystore.KeyStore',
  file=DESCRIPTOR,
  index=0,
  serialized_options=None,
  serialized_start=136,
  serialized_end=244,
  methods=[
  _descriptor.MethodDescriptor(
    name='Get',
    full_name='keystore.KeyStore.Get',
    index=0,
    containing_service=None,
    input_type=_GETREQUEST,
    output_type=_GETRESPONSE,
    serialized_options=None,
  ),
  _descriptor.MethodDescriptor(
    name='Put',
    full_name='keystore.KeyStore.Put',
    index=1,
    containing_service=None,
    input_type=_PUTREQUEST,
    output_type=_EMPTY,
    serialized_options=None,
  ),
])
_sym_db.RegisterServiceDescriptor(_KEYSTORE)

DESCRIPTOR.services_by_name['KeyStore'] = _KEYSTORE

# @@protoc_insertion_point(module_scope)