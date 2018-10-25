import logging

from apicmd import NetAppCommand

from NaServer import NaElement

LOGGER = logging.getLogger(__name__)

class InfoCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'GET.SYS.INFO'

    def execute(self, server, cmd_data_json):
        call = NaElement("system-get-ontapi-version")
        resp, err_resp = self._INVOKE_CHECK(
            server, call, "system-get-ontapi-version")
        if err_resp:
            return err_resp

        majvnum = resp.child_get_int('major-version')
        minvnum = resp.child_get_int('minor-version')
        
        LOGGER.debug("ONTAPI v %d.%d" % (majvnum, minvnum))
        
        call = NaElement("system-get-version")
        resp, err_resp = self._INVOKE_CHECK(
            server, call, "system-get-version")
        if err_resp:
            return err_resp
 
        os_version = resp.child_get_string('version')
        LOGGER.debug("OS Version: %s", os_version)

        return {
            'success' : True, 'errmsg': '', 
            'data': {
                'ontap_major': majvnum,
                'ontap_minor': minvnum,
                'os_version': os_version
            }}

def secondsToText(secs):
    # source: https://gist.github.com/Highstaker/280a09591df4a5fb1363b0bbaf858f0d
    days = secs//86400
    hours = (secs - days*86400)//3600
    minutes = (secs - days*86400 - hours*3600)//60
    seconds = secs - days*86400 - hours*3600 - minutes*60
    days_text = "day{}".format("s" if days!=1 else "")
    hours_text = "hour{}".format("s" if hours!=1 else "")
    minutes_text = "minute{}".format("s" if minutes!=1 else "")
    seconds_text = "second{}".format("s" if seconds!=1 else "")
    result = ", ".join(filter(lambda x: bool(x),[
        "{0} {1}".format(days, days_text) if days else "",
        "{0} {1}".format(hours, hours_text) if hours else "",
        "{0} {1}".format(minutes, minutes_text) if minutes else "",
        "{0} {1}".format(seconds, seconds_text) if seconds else ""
	]))
    return result

class GetNodeCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'SYS.GET.NODE'

    def execute(self, server, cmd_data_json):
        # create the node details query first to see if valid
        qe_ndt = NaElement("node-details-info")
        qpcnt = 0
        if 'name' in cmd_data_json:
            qe_ndt.child_add_string("node", cmd_data_json['name'])
            qpcnt += 1
        if 'uuid' in cmd_data_json:
            qe_ndt.child_add_string("node-uuid", cmd_data_json['uuid'])
            qpcnt += 1

        if qpcnt < 1:
            return self._CREATE_FAIL_RESPONSE(
                'need at least one query parameter, got: '
                + str(cmd_data_json))

        api = NaElement("system-node-get-iter")
        
        xi = NaElement("desired-attributes")
        xi1 = NaElement("node-details-info")
        xi1.child_add_string("node-uuid","<node-uuid>")
        xi1.child_add_string("node","<node>")
        xi1.child_add_string("node-serial-number","<node-serial-number>")
        xi1.child_add_string("node-system-id","<node-system-id>")
        xi1.child_add_string("product-version","<product-version>")
        xi1.child_add_string("is-node-healthy","<is-node-healthy>")
        xi1.child_add_string("node-uptime","<node-uptime>")
        xi.child_add(xi1)
        api.child_add(xi)
                
        qe = NaElement("query")
        qe.child_add(qe_ndt)
        api.child_add(qe)
        
        resp, err_resp = self._INVOKE_CHECK(
            server, api, "system-node-get-iter")
        if err_resp:
            return err_resp

        node_cnt = resp.child_get_int('num-records')
        if not node_cnt or node_cnt > 1:
            # either None or 0 evaluates to False
            return self._CREATE_FAIL_RESPONSE(
                'no nodes or too many found for query: ['
                + str(cmd_data_json) + '] result is: '
                + resp.sprintf())

        if not resp.child_get("attributes-list"):
            return self._CREATE_FAIL_RESPONSE(
                'no node data found in: '
                + resp.sprintf())

        node_detail = resp.child_get("attributes-list").children_get()[0]
        uuid = node_detail.child_get_string("node-uuid")
        name = node_detail.child_get_string("node")
        serial = node_detail.child_get_string("node-serial-number")
        sysid = node_detail.child_get_string("node-system-id")
        version = node_detail.child_get_string("product-version")
        healthy = node_detail.child_get_string("is-node-healthy")
        healthy = True if healthy == 'true' else False
        uptime = node_detail.child_get_int("node-uptime")

        return {
            'success' : True, 'errmsg': '', 
            'data': {
                "name": name,
                "serial": serial,
                "id": sysid,
                "uuid": uuid,
                "version": version,
                "healthy": healthy,
                "uptime": uptime
            }}
