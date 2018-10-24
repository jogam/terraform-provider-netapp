import logging

from apicmd import NetAppCommand

from NaServer import NaElement

LOGGER = logging.getLogger(__name__)

class InfoCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'GET.SYS.INFO'

    def execute(self, server, cmd_data_json):
        api = NaElement("system-get-ontapi-version")
        xo = server.invoke_elem(api)
        if (xo.results_errno() != 0):
            return self.__CREATE_FAIL_RESPONSE(
                '[system-get-ontapi-version] returned: '
                + xo.sprintf())
 
        majvnum = xo.child_get_int('major-version')
        minvnum = xo.child_get_int('minor-version')
        
        LOGGER.debug("ONTAPI v %d.%d" % (majvnum, minvnum))
        
        api1 = NaElement("system-get-version")
        xo1 = server.invoke_elem(api1)
        if (xo1.results_errno() != 0):
            return self.__CREATE_FAIL_RESPONSE(
                '[system-get-version] returned: '
                + xo1.sprintf())
 
        os_version = xo1.child_get_string('version')
        LOGGER.debug("OS Version: %s", os_version)

        return {
            'success' : True, 'errmsg': '', 
            'data': {
                'ontap_major': majvnum,
                'ontap_minor': minvnum,
                'os_version': os_version
            }}