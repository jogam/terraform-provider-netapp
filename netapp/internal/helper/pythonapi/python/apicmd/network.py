import logging

from apicmd import NetAppCommand

from NaServer import NaElement

LOGGER = logging.getLogger(__name__)

class VlanGetCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'NW.VLAN.GET'

    def execute(self, server, cmd_data_json):
        if (
                "parent_name" not in cmd_data_json or
                "vlan_id" not in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                'get vlan request must have parent name'
                + ' and vlan id defined, got: '
                + str(cmd_data_json))

        parent = cmd_data_json["parent_name"]
        vlan_id = cmd_data_json["vlan_id"]
        cmd = "net-vlan-get-iter"

        call = NaElement(cmd)

        qe = NaElement("query")
        qe_vi = NaElement("vlan-info")
        qe_vi.child_add_string("parent-interface", parent)
        qe_vi.child_add_string("vlanid", vlan_id)
        if "node_name" in cmd_data_json:
            qe_vi.child_add_string(
                "node", cmd_data_json["node_name"])
        qe.child_add(qe_vi)
        call.child_add(qe)

        des_attr = NaElement("desired-attributes")
        vi = NaElement("vlan-info")
        vi.child_add_string("interface-name","<interface-name>")
        vi.child_add_string("node","<node>")
        vi.child_add_string("parent-interface","<parent-interface>")
        vi.child_add_string("vlanid","<vlanid>")
        des_attr.child_add(vi)
        call.child_add(des_attr)
        

        resp, err_resp = self._INVOKE_CHECK(
            server, call, 
            cmd + ": " + parent + ":" + vlan_id)
        if err_resp:
            return err_resp

        #LOGGER.debug(resp.sprintf())

        vlan_cnt = self._GET_INT(resp, 'num-records')
        if not vlan_cnt or vlan_cnt < 1:
            # either None or 0 evaluates to False
            return self._CREATE_FAIL_RESPONSE(
                'no vlans or too many found for query: ['
                + str(cmd_data_json) + '] result is: '
                + resp.sprintf())

        if not resp.child_get("attributes-list"):
            return self._CREATE_FAIL_RESPONSE(
                'no vlan data found in: '
                + resp.sprintf())

        vlan_info = resp.child_get("attributes-list").children_get()[0]

        dd = {
            "name": self._GET_STRING(vlan_info, "interface-name"),
            "node_name": self._GET_STRING(vlan_info, "node"),
            "parent_name": self._GET_STRING(vlan_info, "parent-interface"),
            "vlan_id": self._GET_STRING(vlan_info, "vlanid")
        }

        return {
            'success' : True, 'errmsg': '', 'data': dd}

class VlanCreateCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'NW.VLAN.CREATE'

    def execute(self, server, cmd_data_json):
        if (
                "parent_name" not in cmd_data_json or
                "vlan_id" not in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                'create vlan request must have parent name'
                + ' and vlan id defined, got: '
                + str(cmd_data_json))

        parent = cmd_data_json["parent_name"]
        vlan_id = cmd_data_json["vlan_id"]
        cmd = "net-vlan-create"

        call = NaElement(cmd)
        
        vi = NaElement("vlan-info")
        vi.child_add_string("parent-interface", parent)
        vi.child_add_string("vlanid", vlan_id)
        if "node_name" in cmd_data_json:
            vi.child_add_string(
                "node", cmd_data_json["node_name"])
        call.child_add(vi)

        _, err_resp = self._INVOKE_CHECK(
            server, call, 
            cmd + ": " + parent + ":" + vlan_id)
        if err_resp:
            return err_resp

        return self._CREATE_EMPTY_RESPONSE(
            True, "")

class VlanDeleteCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'NW.VLAN.DELETE'

    def execute(self, server, cmd_data_json):
        if (
                "parent_name" not in cmd_data_json or
                "vlan_id" not in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                'delete vlan request must have parent name'
                + ' and vlan id defined, got: '
                + str(cmd_data_json))

        parent = cmd_data_json["parent_name"]
        vlan_id = cmd_data_json["vlan_id"]
        cmd = "net-vlan-delete"

        call = NaElement(cmd)
        
        vi = NaElement("vlan-info")
        vi.child_add_string("parent-interface", parent)
        vi.child_add_string("vlanid", vlan_id)
        if "node_name" in cmd_data_json:
            vi.child_add_string(
                "node", cmd_data_json["node_name"])
        call.child_add(vi)

        _, err_resp = self._INVOKE_CHECK(
            server, call, 
            cmd + ": " + parent + ":" + vlan_id)
        if err_resp:
            return err_resp

        return self._CREATE_EMPTY_RESPONSE(
            True, "")

class IPSpaceGetCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'NW.IPSPACE.GET'

    def execute(self, server, cmd_data_json):
        if not (
            "uuid" in cmd_data_json or
            "name" in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                'get IPSpace request must have uuid'
                + ' defined, got: '
                + str(cmd_data_json))

        cmd = "net-ipspaces-get-iter"

        call = NaElement(cmd)

        qe = NaElement("query")
        qe_ii = NaElement("net-ipspaces-info")

        mark = "unknown"
        if "uuid" in cmd_data_json:
            mark = cmd_data_json["uuid"]
            qe_ii.child_add_string("uuid", mark)
        if "name" in cmd_data_json:
            mark = cmd_data_json["name"]
            qe_ii.child_add_string("ipspace", mark)

        qe.child_add(qe_ii)
        call.child_add(qe)

        des_attr = NaElement("desired-attributes")
        ii = NaElement("net-ipspaces-info")
        ii.child_add_string("ipspace","<ipspace>")
        ii.child_add_string("uuid","<uuid>")

        bcd = NaElement("broadcast-domains")
        bcd.child_add_string("broadcast-domain-name","<broadcast-domain-name>")
        ii.child_add(bcd)

        pts = NaElement("ports")
        pts.child_add_string("net-qualified-port-name","<net-qualified-port-name>")
        ii.child_add(pts)

        vser = NaElement("vservers")
        vser.child_add_string("vserver-name","<vserver-name>")
        ii.child_add(vser)

        des_attr.child_add(ii)
        call.child_add(des_attr)
        
        resp, err_resp = self._INVOKE_CHECK(
            server, call, cmd + ": " + mark)
        if err_resp:
            return err_resp

        vlan_cnt = self._GET_INT(resp, 'num-records')
        if vlan_cnt != 1:
            # too many vlan's found for query
            return self._CREATE_FAIL_RESPONSE(
                'too many ipspaces found for query: ['
                + str(cmd_data_json) + '] result is: '
                + resp.sprintf())

        if not resp.child_get("attributes-list"):
            return self._CREATE_FAIL_RESPONSE(
                'no ipspace data found in: '
                + resp.sprintf())

        ips_info = resp.child_get("attributes-list").children_get()[0]

        dd = {
            "name": self._GET_STRING(ips_info, "ipspace"),
            "uuid": self._GET_STRING(ips_info, "uuid"),
            "bc_domains": self._GET_CONTENT_LIST(ips_info, "broadcast-domains"),
            "ports":  self._GET_CONTENT_LIST(ips_info, "ports"),
            "vservers":  self._GET_CONTENT_LIST(ips_info, "vservers")
        }

        return {
            'success' : True, 'errmsg': '', 'data': dd}

class IPSpaceCreateCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'NW.IPSPACE.CREATE'

    def execute(self, server, cmd_data_json):
        if "name" not in cmd_data_json:
            return self._CREATE_FAIL_RESPONSE(
                'create IPSpace request must have name'
                + ' defined, got: '
                + str(cmd_data_json))

        name = cmd_data_json["name"]
        cmd = "net-ipspaces-create"

        call = NaElement(cmd)
        call.child_add_string("ipspace", name)
        call.child_add_string("return-record", "true")

        resp, err_resp = self._INVOKE_CHECK(
            server, call, cmd + ": " + name)
        if err_resp:
            return err_resp

        if not (
            resp.child_get("result") and
            resp.child_get(
                "result").child_get("net-ipspaces-info")):
            return self._CREATE_FAIL_RESPONSE(
                'no ipspace info received from create, got: '
                + resp.sprintf())

        ipspace_info = resp.child_get(
            "result").child_get("net-ipspaces-info")
        dd = {
            "uuid": self._GET_STRING(ipspace_info, "uuid")
        }

        return {
             'success' : True, 'errmsg': '', 'data': dd}

class IpSpaceDeleteCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'NW.IPSPACE.DELETE'

    def execute(self, server, cmd_data_json):
        if "name" not in cmd_data_json:
            return self._CREATE_FAIL_RESPONSE(
                'delete ipspace must have name defined, got: '
                + str(cmd_data_json))

        name = cmd_data_json["name"]
        cmd = "net-ipspaces-destroy"

        call = NaElement(cmd)
        
        call.child_add_string("ipspace", name)

        _, err_resp = self._INVOKE_CHECK(
            server, call, cmd + ": " + name)
        if err_resp:
            return err_resp

        return self._CREATE_EMPTY_RESPONSE(
            True, "")

class IpSpaceUpdateCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'NW.IPSPACE.UPDATE'

    def execute(self, server, cmd_data_json):
        if (
                "name" not in cmd_data_json or
                "new_name" not in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                'update/rename ipspace must have name'
                + ' and new_name defined, got: '
                + str(cmd_data_json))

        name = cmd_data_json["name"]
        new_name = cmd_data_json["new_name"]
        cmd = "net-ipspaces-rename"

        call = NaElement(cmd)
        
        call.child_add_string("ipspace", name)
        call.child_add_string("new-name", new_name)

        _, err_resp = self._INVOKE_CHECK(
            server, call, cmd + ": " 
            + name + ' to ' + new_name)
        if err_resp:
            return err_resp

        return self._CREATE_EMPTY_RESPONSE(
            True, "")

class BcDomainGetCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return "NW.BRCDOM.GET"

    def execute(self, server, cmd_data_json):
        if "name" not in cmd_data_json:
            return self._CREATE_FAIL_RESPONSE(
                "get broadcast domain commands must"
                + " have name defined, got: "
                + str(cmd_data_json))

        name = cmd_data_json["name"]
        cmd = "net-port-broadcast-domain-get-iter"

        call = NaElement(cmd)

        qe = NaElement("query")
        qe_bcdi = NaElement("net-port-broadcast-domain-info")
        qe_bcdi.child_add_string("broadcast-domain", name)
        qe.child_add(qe_bcdi)
        call.child_add(qe)

        dattr = NaElement("desired-attributes")
        bcdi = NaElement("net-port-broadcast-domain-info")
        bcdi.child_add_string("broadcast-domain","<broadcast-domain>")
        bcdi.child_add_string("ipspace","<ipspace>")
        bcdi.child_add_string("mtu","<mtu>")
        bcdi.child_add_string("port-update-status-combined","<port-update-status-combined>")

        ports = NaElement("ports")
        pi = NaElement("port-info")
        pi.child_add_string("port","<port>")
        pi.child_add_string("port-update-status","<port-update-status>")
        pi.child_add_string("port-update-status-details","<port-update-status-details>")        
        ports.child_add(pi)
        bcdi.child_add(ports)

        fog = NaElement("failover-groups")
        fog.child_add_string("failover-group","<failover-group>")
        bcdi.child_add(fog)

        subs = NaElement("subnet-names")
        subs.child_add_string("subnet-name","<subnet-name>")        
        bcdi.child_add(subs)

        dattr.child_add(bcdi)

        call.child_add(dattr)

        resp, err_resp = self._INVOKE_CHECK(
            server, call, cmd + ": " + name)
        if err_resp:
            return err_resp

        LOGGER.debug(resp.sprintf())

        bcd_cnt = self._GET_INT(resp, 'num-records')
        if bcd_cnt != 1:
            # too many bc domains found for query
            return self._CREATE_FAIL_RESPONSE(
                'too many broadcast domains found for'
                + ' query: [' + str(cmd_data_json) + '] result is: '
                + resp.sprintf())

        if not resp.child_get("attributes-list"):
            return self._CREATE_FAIL_RESPONSE(
                'no broadcast domain info data found in: '
                + resp.sprintf())

        bcd_info = resp.child_get("attributes-list").children_get()[0]

        dd = {
            "name": self._GET_STRING(bcd_info, "broadcast-domain"),
            "mtu": self._GET_STRING(bcd_info, "mtu"),
            "ipspace": self._GET_STRING(bcd_info, "ipspace"),
            "update_status": self._GET_STRING(bcd_info, "port-update-status-combined"),
            "ports": [],
            "failovergrps":  self._GET_CONTENT_LIST(bcd_info, "failover-groups"),
            "subnets":  self._GET_CONTENT_LIST(bcd_info, "subnet-names")
        }

        if bcd_info.child_get("ports"):
            # port info data available, process
            for port_info in bcd_info.child_get("ports").children_get():
                dd['ports'].append({
                    "name": self._GET_STRING(port_info, "port"),
                    "update_status": self._GET_STRING(port_info, "port-update-status"),
                    "status_detail": self._GET_STRING(port_info, "port-update-status-details")
                })

        return {
            'success' : True, 'errmsg': '', 'data': dd}

class BcDomainStatusCommand(NetAppCommand):
 
    @classmethod
    def get_name(cls):
        return "NW.BRCDOM.STATUS"

    def execute(self, server, cmd_data_json):
        if "name" not in cmd_data_json:
            return self._CREATE_FAIL_RESPONSE(
                "broadcast domain status commands must"
                + " have name defined, got: "
                + str(cmd_data_json))

        name = cmd_data_json["name"]
        cmd = "net-port-broadcast-domain-get-iter"

        call = NaElement(cmd)

        qe = NaElement("query")
        qe_bcdi = NaElement("net-port-broadcast-domain-info")
        qe_bcdi.child_add_string("broadcast-domain", name)
        qe.child_add(qe_bcdi)
        call.child_add(qe)

        dattr = NaElement("desired-attributes")
        bcdi = NaElement("net-port-broadcast-domain-info")
        bcdi.child_add_string("port-update-status-combined","<port-update-status-combined>")
        dattr.child_add(bcdi)

        call.child_add(dattr)

        resp, err_resp = self._INVOKE_CHECK(
            server, call, cmd + ": " + name)
        if err_resp:
            return err_resp

        LOGGER.debug(resp.sprintf())

        bcd_cnt = self._GET_INT(resp, 'num-records')
        if bcd_cnt != 1:
            # too many bc domains found for query
            return self._CREATE_FAIL_RESPONSE(
                'too many broadcast domains found for'
                + ' query: [' + str(cmd_data_json) + '] result is: '
                + resp.sprintf())

        if not resp.child_get("attributes-list"):
            return self._CREATE_FAIL_RESPONSE(
                'no broadcast domain info data found in: '
                + resp.sprintf())

        bcd_info = resp.child_get("attributes-list").children_get()[0]

        dd = {
            "update_status": self._GET_STRING(bcd_info, "port-update-status-combined"),
        }

        return {
            'success' : True, 'errmsg': '', 'data': dd}

class BcDomainCreateCommand(NetAppCommand):
 
    @classmethod
    def get_name(cls):
        return "NW.BRCDOM.CREATE"

    def execute(self, server, cmd_data_json):
        if (
                "name" not in cmd_data_json or
                "mtu"  not in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                "broadcast domain create commands must"
                + " have name and mtu defined, got: "
                + str(cmd_data_json))

        name = cmd_data_json["name"]
        mtu = cmd_data_json["mtu"]
        cmd = "net-port-broadcast-domain-create"

        call = NaElement(cmd)

        call.child_add_string("broadcast-domain", name)
        call.child_add_string("mtu", mtu)

        if "ipspace" in cmd_data_json:
            call.child_add_string("ipspace", cmd_data_json["ipspace"])

        if "ports" in cmd_data_json:
            ports = NaElement("ports")

            for port_name in cmd_data_json["ports"]:
                ports.child_add_string(
                    "net-qualified-port-name", port_name)

            call.child_add(ports)

        resp, err_resp = self._INVOKE_CHECK(
            server, call, cmd + ": " + name)
        if err_resp:
            return err_resp

        LOGGER.debug(resp.sprintf())

        dd = {
            "update_status": self._GET_STRING(resp, "port-update-status-combined"),
        }

        return {
            'success' : True, 'errmsg': '', 'data': dd}

class BcDomainDeleteCommand(NetAppCommand):
 
    @classmethod
    def get_name(cls):
        return "NW.BRCDOM.DELETE"

    def execute(self, server, cmd_data_json):
        if "name" not in cmd_data_json:
            return self._CREATE_FAIL_RESPONSE(
                "broadcast domain delete commands must"
                + " have name defined, got: "
                + str(cmd_data_json))

        name = cmd_data_json["name"]
        cmd = "net-port-broadcast-domain-destroy"

        call = NaElement(cmd)

        call.child_add_string("broadcast-domain", name)

        resp, err_resp = self._INVOKE_CHECK(
            server, call, cmd + ": " + name)
        if err_resp:
            return err_resp

        LOGGER.debug(resp.sprintf())

        dd = {
            "update_status": self._GET_STRING(resp, "port-update-status-combined"),
        }

        return {
            'success' : True, 'errmsg': '', 'data': dd}