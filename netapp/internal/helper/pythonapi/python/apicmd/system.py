import logging

from apicmd import NetAppCommand

from NaServer import NaElement

LOGGER = logging.getLogger(__name__)

class InfoCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'SYS.INFO.GET'

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

class NodeGetCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'SYS.NODE.GET'

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

class PortGetCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'SYS.PORT.GET'

    def execute(self, server, cmd_data_json):
        if (
                "node" not in cmd_data_json or
                "port" not in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                'get port request must have node and port defined, got: '
                + str(cmd_data_json))

        node = cmd_data_json["node"]
        port = cmd_data_json["port"]
        cmd = "net-port-get-iter"

        call = NaElement(cmd)

        qe = NaElement("query")
        qe_npi = NaElement("net-port-info")
        qe_npi.child_add_string("node", node)
        qe_npi.child_add_string("port", port)
        qe.child_add(qe_npi)
        call.child_add(qe)

        des_attr = NaElement("desired-attributes")
        npi = NaElement("net-port-info")
        npi.child_add_string("node","<node>")
        npi.child_add_string("port","<port>")
        npi.child_add_string("autorevert-delay","<autorevert-delay>")
        npi.child_add_string("ignore-health-status","<ignore-health-status>")
        npi.child_add_string("ipspace","<ipspace>")
        npi.child_add_string("role","<role>")
        npi.child_add_string("is-administrative-up","<is-administrative-up>")
        npi.child_add_string("mtu-admin","<mtu-admin>")
        npi.child_add_string("is-administrative-auto-negotiate","<is-administrative-auto-negotiate>")
        npi.child_add_string("administrative-speed","<administrative-speed>")
        npi.child_add_string("administrative-duplex","<administrative-duplex>")
        npi.child_add_string("administrative-flowcontrol","<administrative-flowcontrol>")
        npi.child_add_string("link-status","<link-status>")
        npi.child_add_string("health-status","<health-status>")
        npi.child_add_string("mac-address","<mac-address>")
        npi.child_add_string("broadcast-domain","<broadcast-domain>")
        npi.child_add_string("mtu","<mtu>")
        npi.child_add_string("is-operational-auto-negotiate","<is-operational-auto-negotiate>")
        npi.child_add_string("operational-speed","<operational-speed>")
        npi.child_add_string("operational-duplex","<operational-duplex>")
        npi.child_add_string("operational-flowcontrol","<operational-flowcontrol>")
        des_attr.child_add(npi)
        call.child_add(des_attr)
        

        resp, err_resp = self._INVOKE_CHECK(
            server, call, 
            cmd + ": " + node + ":" + port)
        if err_resp:
            return err_resp

        #LOGGER.debug(resp.sprintf())

        port_cnt = resp.child_get_int('num-records')
        if not port_cnt or port_cnt > 1:
            # either None or 0 evaluates to False
            return self._CREATE_FAIL_RESPONSE(
                'no ports or too many found for query: ['
                + str(cmd_data_json) + '] result is: '
                + resp.sprintf())

        if not resp.child_get("attributes-list"):
            return self._CREATE_FAIL_RESPONSE(
                'no port data found in: '
                + resp.sprintf())

        port_info = resp.child_get("attributes-list").children_get()[0]

        dd = {
            "node": self._GET_STRING(port_info, "node"),
            "port": self._GET_STRING(port_info, "port"),

            "auto_rev_delay": self._GET_INT(port_info, "autorevert-delay"),
            "ignr_health": self._GET_BOOL(port_info, "ignore-health-status"),
            "ipspace": self._GET_STRING(port_info, "ipspace"),
            "role": self._GET_STRING(port_info, "role"),

            "admin_up": self._GET_BOOL(port_info, "is-administrative-up"),
            "admin_mtu": self._GET_INT(port_info, "mtu-admin"),
            "admin_auto": self._GET_BOOL(port_info, "is-administrative-auto-negotiate"),
            "admin_speed": self._GET_STRING(port_info, "administrative-speed"),
            "admin_duplex": self._GET_STRING(port_info, "administrative-duplex"),
            "admin_flow": self._GET_STRING(port_info, "administrative-flowcontrol"),

            "status": self._GET_STRING(port_info, "link-status"),
            "health": self._GET_STRING(port_info, "health-status"),
            "mac": self._GET_STRING(port_info, "mac-address"),
            "broadcast_domain": self._GET_STRING(port_info, "broadcast-domain"),
            "mtu": self._GET_INT(port_info, "mtu"),
            "auto": self._GET_BOOL(port_info, "is-operational-auto-negotiate"),
            "speed": self._GET_STRING(port_info, "operational-speed"),
            "duplex": self._GET_STRING(port_info, "operational-duplex"),
            "flow": self._GET_STRING(port_info, "operational-flowcontrol")
        }

        return {
            'success' : True, 'errmsg': '', 'data': dd}

class PortModifyCommand(NetAppCommand):
    __cmd_mapping = {
        "duplex": "administrative-duplex",
        "flow": "administrative-flowcontrol",
        "speed": "administrative-speed",
        "auto_rev_delay": "autorevert-delay",
        "ignr_health": "ignore-health-status",
        "ipspace":"ipspace",
        "auto": "is-administrative-auto-negotiate",
        "up": "is-administrative-up",
        "mtu": "mtu",
        "role": "role"
    }

    @classmethod
    def get_name(cls):
        return 'SYS.PORT.MODIFY'

    def execute(self, server, cmd_data_json):
        if (
                "node" not in cmd_data_json or
                "port" not in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                'modify port request must have node and port defined, got: '
                + str(cmd_data_json))

        node = cmd_data_json["node"]
        port = cmd_data_json["port"]
        cmd = "net-port-modify"

        call = NaElement(cmd)
        call.child_add_string("node", node)
        call.child_add_string("port", port)

        for cmd_data_key, netapp_cmd_str in self.__cmd_mapping.items():
            if cmd_data_key in cmd_data_json:
                call.child_add_string(
                    netapp_cmd_str,
                    str(cmd_data_json[cmd_data_key]))

        resp, err_resp = self._INVOKE_CHECK(
            server, call, 
            cmd + ": " + node + ":" + port)
        if err_resp:
            return err_resp

        LOGGER.debug(resp.sprintf())

        return self._CREATE_EMPTY_RESPONSE(
            True, "")
