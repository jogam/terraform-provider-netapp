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

        node_cnt = self._GET_INT(resp, 'num-records')
        if node_cnt != 1:
            # too many nodes found for query
            return self._CREATE_FAIL_RESPONSE(
                'too many nodes found for query: ['
                + str(cmd_data_json) + '] result is: '
                + resp.sprintf())

        if not resp.child_get("attributes-list"):
            return self._CREATE_FAIL_RESPONSE(
                'no node data found in: '
                + resp.sprintf())

        node_detail = resp.child_get("attributes-list").children_get()[0]
        uuid = self._GET_STRING(node_detail, "node-uuid")
        name = self._GET_STRING(node_detail, "node")
        serial = self._GET_STRING(node_detail, "node-serial-number")
        sysid = self._GET_STRING(node_detail, "node-system-id")
        version = self._GET_STRING(node_detail, "product-version")
        healthy = self._GET_BOOL(node_detail, "is-node-healthy")
        uptime = self._GET_INT(node_detail, "node-uptime")

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
        npi.child_add_string("port-type","<port-type>")
        npi.child_add_string("vlan-id","<vlan-id>")
        npi.child_add_string("vlan-node","<vlan-node>")
        npi.child_add_string("vlan-port","<vlan-port>")

        des_attr.child_add(npi)
        call.child_add(des_attr)
        

        resp, err_resp = self._INVOKE_CHECK(
            server, call, 
            cmd + ": " + node + ":" + port)
        if err_resp:
            return err_resp

        #LOGGER.debug(resp.sprintf())

        port_cnt = self._GET_INT(resp, 'num-records')
        if port_cnt != 1:
            # too many ports received for query
            return self._CREATE_FAIL_RESPONSE(
                'too many ports found for query: ['
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

            "auto_rev_delay": self._GET_STRING(port_info, "autorevert-delay"),
            "ignr_health": self._GET_STRING(port_info, "ignore-health-status"),
            "ipspace": self._GET_STRING(port_info, "ipspace"),
            "role": self._GET_STRING(port_info, "role"),

            "admin_up": self._GET_STRING(port_info, "is-administrative-up"),
            "admin_mtu": self._GET_STRING(port_info, "mtu-admin"),
            "admin_auto": self._GET_STRING(port_info, "is-administrative-auto-negotiate"),
            "admin_speed": self._GET_STRING(port_info, "administrative-speed"),
            "admin_duplex": self._GET_STRING(port_info, "administrative-duplex"),
            "admin_flow": self._GET_STRING(port_info, "administrative-flowcontrol"),

            "status": self._GET_STRING(port_info, "link-status"),
            "health": self._GET_STRING(port_info, "health-status"),
            "mac": self._GET_STRING(port_info, "mac-address"),
            "broadcast_domain": self._GET_STRING(port_info, "broadcast-domain"),
            "mtu": self._GET_STRING(port_info, "mtu"),
            "auto": self._GET_STRING(port_info, "is-operational-auto-negotiate"),
            "speed": self._GET_STRING(port_info, "operational-speed"),
            "duplex": self._GET_STRING(port_info, "operational-duplex"),
            "flow": self._GET_STRING(port_info, "operational-flowcontrol"),

            "type": self._GET_STRING(port_info, "port-type"),

	        "vlan_id": self._GET_STRING(port_info, "vlan-id"),
	        "vlan_node": self._GET_STRING(port_info, "vlan-node"),
	        "vlan_port": self._GET_STRING(port_info, "vlan-port")
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
                    cmd_data_json[cmd_data_key])

        _, err_resp = self._INVOKE_CHECK(
            server, call, 
            cmd + ": " + node + ":" + port)
        if err_resp:
            return err_resp

        return self._CREATE_EMPTY_RESPONSE(
            True, "")

class AggrGetCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'SYS.AGGR.GET'

    def execute(self, server, cmd_data_json):
        if (
                "name" not in cmd_data_json and
                "uuid" not in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                'get aggr request must have either '
                + 'name or uuid defined, got: '
                + str(cmd_data_json))

        cmd = "aggr-get-iter"

        call = NaElement(cmd)

        qe = NaElement("query")
        qi_aa = NaElement("aggr-attributes")

        for key in ["name", "uuid"]:
            if key in cmd_data_json:
                qi_aa.child_add_string(
                    "aggregate-" + key, cmd_data_json[key])

        if "nodes" in cmd_data_json:
            qi_nodes = NaElement("nodes")

            for node_name in cmd_data_json["nodes"]:
                qi_nodes.child_add_string("node-name", node_name)

            qi_aa.child_add(qi_nodes)

        qe.child_add(qi_aa)
        call.child_add(qe)

        des_attr = NaElement("desired-attributes")

        agg_attr = NaElement("aggr-attributes")
        agg_attr.child_add_string("aggregate-name","<aggregate-name>")
        agg_attr.child_add_string("aggregate-uuid","<aggregate-uuid>")

        agg_spc = NaElement("aggr-space-attributes")
        agg_spc.child_add_string("percent-used-capacity","<percent-used-capacity>")
        agg_spc.child_add_string("physical-used-percent","<physical-used-percent>")
        agg_spc.child_add_string("size-available","<size-available>")
        agg_spc.child_add_string("size-total","<size-total>")
        agg_spc.child_add_string("size-used","<size-used>")
        agg_spc.child_add_string("total-reserved-space","<total-reserved-space>")
        agg_attr.child_add(agg_spc)

        agg_vcnt = NaElement("aggr-volume-count-attributes")
        agg_vcnt.child_add_string("flexvol-count","<flexvol-count>")
        agg_attr.child_add(agg_vcnt)

        des_attr.child_add(agg_attr)
        call.child_add(des_attr)

        resp, err_resp = self._INVOKE_CHECK(
            server, call, cmd + ": <-- "
            + str(cmd_data_json))
        if err_resp:
            return err_resp

        agg_cnt = self._GET_INT(resp, 'num-records')
        if agg_cnt != 1:
            # too many aggregates found for query
            return self._CREATE_FAIL_RESPONSE(
                'too many aggregates found for'
                + ' query: [' + str(cmd_data_json) 
                + '] result is: '
                + resp.sprintf())

        if not resp.child_get("attributes-list"):
            return self._CREATE_FAIL_RESPONSE(
                'no aggregate info data found in: '
                + resp.sprintf())

        agg_info = resp.child_get("attributes-list").children_get()[0]
        dd = {
            "name": self._GET_STRING(agg_info, "aggregate-name"),
            "uuid": self._GET_STRING(agg_info, "aggregate-uuid"),
        }

        agg_si = agg_info.child_get("aggr-space-attributes")
        if agg_si:
            dd["pct_used_cap"] = self._GET_INT(agg_si, "percent-used-capacity")
            dd["pct_used_phys"] = self._GET_INT(agg_si, "physical-used-percent")
            dd["size_avail"] = self._GET_INT(agg_si, "size-available")
            dd["size_total"] = self._GET_INT(agg_si, "size-total")
            dd["size_used"] = self._GET_INT(agg_si, "size-used")
            dd["size_reserve"] = self._GET_INT(agg_si, "total-reserved-space")

        add_vco = agg_info.child_get("aggr-volume-count-attributes")
        if add_vco:
            dd["flexvol_cnt"] = self._GET_INT(add_vco, "flexvol-count")

        return {
            'success' : True, 'errmsg': '', 'data': dd}