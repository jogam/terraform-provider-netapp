import logging

from apicmd import NetAppCommand

LOGGER = logging.getLogger(__name__)

class KeyValueCommand(NetAppCommand):

    def __init__(self):
        self.keyvaluemap = {}

    @classmethod
    def get_name(cls):
        return 'TEST.KEYVALUE'

    def execute(self, server, cmd_data_json):
        key = cmd_data_json.get('key', 'KEYERROR')
        value = cmd_data_json.get('value', 'VALUEERROR')
        write = cmd_data_json.get('write', 'WRITEERROR')

        LOGGER.debug(
            'key: %s; value: %s, write: %s',
            key, value, write)

        if 'ERROR' in key or 'ERROR' in value or type(write) is not bool:
            return {
                'success': False, 'data': {},
                'errmsg': 'cmd data error, got: ' + str(cmd_data_json)
            }

        modified = False
        if write:
            self.keyvaluemap[key] = value
            modified = True
        else:
            if not key in self.keyvaluemap:
                return {
                    'success': False, 'data': {},
                    'errmsg': 'no value for key: ' + key
                }
            
            value = self.keyvaluemap[key]

        return {
            'success' : True, 'errmsg': '', 
            'data': {
                'key': key, 'value': value, 'modified': modified}
        }