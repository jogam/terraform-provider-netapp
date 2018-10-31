import sys
import os
import logging
import json

# append the NetApp SDK python library to path
sys.path.append(
    os.path.join(
        os.environ.get('NETAPP_MSDK_ROOT_PATH', 'FAULT'),
        'lib', 'python', 'NetApp'))

from NaServer import NaServer

import importlib
import pkgutil

LOGGER = logging.getLogger(__name__)

def import_submodules(package, recursive=True):
    """ Import all submodules of a module, recursively, including subpackages
    [Source]: https://stackoverflow.com/a/25562415
    Usage:

        # from main.py, as per the OP's project structure
        import scripts
        import_submodules(scripts)

        # Alternatively, from scripts.__init__.py
        import_submodules(__name__)

    :param package: package (name or actual module)
    :type package: str | module
    :rtype: dict[str, types.ModuleType]
    """
    LOGGER.debug('import submodules for: %s', package)
    if isinstance(package, str):
        package = importlib.import_module(package)
    results = {}
    for loader, name, is_pkg in pkgutil.walk_packages(package.__path__):
        full_name = package.__name__ + '.' + name
        LOGGER.debug('importing module [%s]', full_name)
        results[full_name] = importlib.import_module(full_name)
        if recursive and is_pkg:
            results.update(import_submodules(full_name))
    return results

class NetAppCommand(object):
    available_implementations = {}

    def __init_subclass__(cls, **kwargs):
        super().__init_subclass__(**kwargs)
        LOGGER.debug(
            '[%s] adding command: %s',
            cls.__name__, cls.get_name())
        cls.__ADD_CMD(cls)

    @classmethod
    def get_name(cls):
        raise NotImplementedError('must be implemented by subclass')

    @staticmethod
    def __ADD_CMD(clazz):
        # create new Command instance and add to mapping
        NetAppCommand.available_implementations[clazz.get_name()] = clazz()
        LOGGER.debug(
            'available commands: %s',
            NetAppCommand.available_implementations.keys())

    @staticmethod
    def GET_CMD(cmd_name):
        if not cmd_name in NetAppCommand.available_implementations:
            raise NotImplementedError('no command with name: ' + cmd_name)

        return NetAppCommand.available_implementations[cmd_name]

    @staticmethod
    def _GET_STRING(na_elem, key):
        return na_elem.child_get_string(key)

    @staticmethod
    def _GET_BOOL(na_elem, key):
        value = NetAppCommand._GET_STRING(na_elem, key)
        return True if value == 'true' else False

    @staticmethod
    def _GET_INT(na_elem, key):
        value = NetAppCommand._GET_STRING(na_elem, key)
        if not value:
            return -1
            
        return int(value) if value.isdigit() else -1

    @staticmethod
    def _GET_CONTENT(na_elem):
        return na_elem.element['content']

    @staticmethod
    def _GET_CONTENT_LIST(na_elem, key):
        values = []
        if na_elem.child_get(key):
            for v_elem in na_elem.child_get(key).children_get():
                values.append(NetAppCommand._GET_CONTENT(v_elem))

        return values
        

    @staticmethod
    def _CREATE_EMPTY_RESPONSE(success, msg):
        return { 
            'success': success, 'errmsg': msg,
            'data': {'dummy': 1} }

    @staticmethod
    def _CREATE_FAIL_RESPONSE(msg):
        return { 'success': False, 'errmsg': msg, 'data': {} }

    @staticmethod
    def _INVOKE_CHECK(server, request, cmd_name):
        response = server.invoke_elem(request)
        err_resp = None
        if response.results_errno() is not 0:
            err_resp = NetAppCommand._CREATE_FAIL_RESPONSE(
                '[' + cmd_name + '] returned: '
                + response.sprintf())

        return response, err_resp

    def execute(self, server, cmd_data_json):
        '''
        execute the command

        :param NaServer server: 
            the NetApp server to use for invoking elements
        :param json cmd_data_json:
            the command input data as received via gRPC
        :return json:
            the command execution result, must contain:
            {
                'success': bool (True: all good),
                'errmsg': string should be empty string if all good
                'data': the command exec result {}
            }
        '''
        raise NotImplementedError('must be implemented by subclass')


# import all commands defined in sub modules
# NOTE: must be done after NetAppCommand definition,
#       otherwise abstract class does not exist!
import_submodules(__name__)


API_ENCODING = 'utf-8'
API_CONNECT_CMD = "SYS.CONNECT"

class NetAppCommandExecutor(object):

    def __init__(self):
        super(NetAppCommandExecutor, self).__init__()
        self.connected  = False

        self.ontap_major_version = 1
        self.ontap_minor_version = 2
        self.os_version = 'NO-INITIALIZED'

        # parameters required for server creation
        self.host = ''
        self.user = ''
        self.pwd = ''

        # these are static for now
        self.server_type = 'FILER'
        self.transport_type = 'HTTP'
        self.server_port = 80
        self.connect_style = 'LOGIN'

    def __create_server(self, testing_active):
        if testing_active:
            return None

        s = NaServer(
                self.host,
                self.ontap_major_version,
                self.ontap_minor_version)
        s.set_server_type(self.server_type)
        s.set_transport_type(self.transport_type)
        s.set_port(self.server_port)
        s.set_style(self.connect_style)
        s.set_admin_user(self.user, self.pwd)
    
        return s

    @staticmethod
    def __GET_COMMAND(name):
        cmd = None
        try:
            cmd = NetAppCommand.GET_CMD(name)
        except NotImplementedError as err:
            LOGGER.error('executor get-command returned: %s', err)

        return cmd

    @staticmethod
    def __BYTES_TO_JSON(byte_data):
        json_str = byte_data.decode(API_ENCODING).replace("'", '"')
        LOGGER.debug('decoded byte data: %s', json_str)
        return json.loads(json_str)

    @staticmethod
    def __JSON_TO_BYTES(json_data):
        json_str = json.dumps(json_data)
        LOGGER.debug('dumped json data: %s', json_str)
        return json_str.encode(API_ENCODING)

    @staticmethod
    def __CREATE_FAIL_RETVAL(errmsg):
        return False, errmsg, b''

    def execute(self, cmd_name, cmd_byte_data):
        '''
        execute a NetApp API command

        :param string cmd_name:
            name of the command to execute
        :param bytes cmd_byte_data: 
            the command input data to use

        :return: succ, errmsg, resp_data
            :param bool succ:
                True if command executed successful
            :param string errmsg:
                error message if succ == False
            :param bytes resp_data:
                command respond data as bytes
        '''
        # create json from command data bytes
        cmd_data = self.__BYTES_TO_JSON(cmd_byte_data)

        connect_active = False
        # NOTE: that might be a little somewhat special...
        testing_active = cmd_name.startswith('TEST.')

        if cmd_name == API_CONNECT_CMD:
            # get command data
            host = cmd_data.get('host', 'HOST-ERROR')
            user = cmd_data.get('user', 'USER-ERROR')
            pwd = cmd_data.get('pwd', 'PWD-ERROR')

            # check for changes and if already connected
            if (
                    host == self.host and
                    user == self.user and
                    pwd == self.pwd and
                    self.connected):

                # already connected and all setup, just return data
                res_data_json = {}
                res_data_json['version_ontap'] = (
                    str(self.ontap_major_version) + '.'
                    + str(self.ontap_minor_version))
                res_data_json['version_os'] = self.os_version

                res_bytes = self.__JSON_TO_BYTES(res_data_json)

                return True, "", res_bytes

            # not connected yet, store data and connect
            self.host = host
            self.user = user
            self.pwd = pwd
            
            cmd_name = 'SYS.INFO.GET'
            connect_active = True

        cmd = self.__GET_COMMAND(cmd_name)
        if not cmd:
            return self.__CREATE_FAIL_RETVAL(
                'could not get command: ' + cmd_name)

        if not(self.connected or connect_active or testing_active):
            # API is not connected yet and connect command is not active
            return self.__CREATE_FAIL_RETVAL(
                'API not connected, call Connect() first')

        cmd_res_dict = cmd.execute(
            self.__create_server(testing_active), cmd_data)

        if not cmd_res_dict:
            LOGGER.error(
                'cmd [%s] did not return response, got %s',
                cmd_name, cmd_res_dict)
            return self.__CREATE_FAIL_RETVAL(
                'cmd [' + cmd_name + '] did not return response')

        res_err_msg = cmd_res_dict.get('errmsg', 'ERRMSG missing')
        res_success = cmd_res_dict.get('success', False)
        if not res_success:
            LOGGER.debug('unsuccessful cmd exec: %s', cmd_res_dict)
            if connect_active:
                # failed to connect...
                self.connected = False
                return self.__CREATE_FAIL_RETVAL(
                    'API connect failed with: ' + res_err_msg )
            
            return self.__CREATE_FAIL_RETVAL(
                'failed cmd [' + cmd_name + '] with ' + res_err_msg)
        
        if len(res_err_msg) > 0:
            LOGGER.warn(
                'cmd [%s] has error msg: %s',
                cmd_name, res_err_msg)

        res_data_json = cmd_res_dict.get('data', {})
        if not res_data_json:
            LOGGER.error(
                'cmd [%s] returned empty data: %s',
                cmd_name, res_data_json)
            return (
                False, 
                'cmd [' + cmd_name + '] no data, with: ' + res_err_msg, 
                b''
            )

        if connect_active:
            self.connected = True
            self.ontap_major_version = res_data_json.pop('ontap_major')
            self.ontap_minor_version = res_data_json.pop('ontap_minor')
            self.os_version = res_data_json.pop('os_version')

            res_data_json['version_ontap'] = (
                str(self.ontap_major_version) + '.'
                + str(self.ontap_minor_version))
            res_data_json['version_os'] = self.os_version

        res_bytes = self.__JSON_TO_BYTES(res_data_json)

        return res_success, res_err_msg, res_bytes
