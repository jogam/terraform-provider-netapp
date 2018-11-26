import logging

from apicmd import NetAppCommand

from NaServer import NaElement

LOGGER = logging.getLogger(__name__)

class SvmGetCommand(NetAppCommand):

    @classmethod
    def get_name(cls):
        return 'SVM.GET'

    def execute(self, server, cmd_data_json):
        if (
                "name" not in cmd_data_json and
                "uuid" not in cmd_data_json):
            return self._CREATE_FAIL_RESPONSE(
                'get SVM request must have name'
                + ' or uuid defined, got: '
                + str(cmd_data_json))

        cmd = "vserver-get-iter"
        call = NaElement(cmd)

        qe = NaElement("query")
        qe_vsi = NaElement("vserver-info")
        if "name" in cmd_data_json:
            qe_vsi.child_add_string(
                "vserver-name",
                cmd_data_json["name"])
        elif "uuid" in cmd_data_json:
            qe_vsi.child_add_string(
                "uuid", cmd_data_json["uuid"])
        qe.child_add(qe_vsi)
        call.child_add(qe)

        des_attr = NaElement("desired-attributes")
        vsi = NaElement("vserver-info")

        for prot_attr in ["allowed-protocols", "disallowed-protocols"]:
            alp = NaElement(prot_attr)
            alp.child_add_string("protocol","<protocol>")
            vsi.child_add(alp)

        vsi.child_add_string("ipspace","<ipspace>")
        vsi.child_add_string("is-config-locked-for-changes","<is-config-locked-for-changes>")
        vsi.child_add_string("operational-state","<operational-state>")
        vsi.child_add_string("operational-state-stopped-reason","<operational-state-stopped-reason>")
        vsi.child_add_string("root-volume","<root-volume>")
        vsi.child_add_string("root-volume-aggregate","<root-volume-aggregate>")
        vsi.child_add_string("root-volume-security-style","<root-volume-security-style>")
        vsi.child_add_string("state","<state>")
        vsi.child_add_string("uuid","<uuid>")
        vsi.child_add_string("volume-delete-retention-hours","<volume-delete-retention-hours>")
        vsi.child_add_string("vserver-name","<vserver-name>")

        des_attr.child_add(vsi)
        call.child_add(des_attr)

        resp, err_resp = self._INVOKE_CHECK(
            server, call, 
            cmd + ": " + str(cmd_data_json))
        if err_resp:
            return err_resp

        LOGGER.debug(resp.sprintf())

        svm_cnt = self._GET_INT(resp, 'num-records')
        if svm_cnt != 1:
            # not exactly one svm received for query
            return self._CREATE_FAIL_RESPONSE(
                'not exactly one SVM received for: ['
                + str(cmd_data_json) + '] result is: '
                + resp.sprintf())

        if not resp.child_get("attributes-list"):
            return self._CREATE_FAIL_RESPONSE(
                'no svm data found in: '
                + resp.sprintf())

        svm_info = resp.child_get("attributes-list").children_get()[0]

        dd = {
            "name": self._GET_STRING(svm_info, "vserver-name"),
            "uuid": self._GET_STRING(svm_info, "uuid"),
            "ipspace": self._GET_STRING(svm_info, "ipspace"),
            "root_aggr": self._GET_STRING(svm_info, "root-volume-aggregate"),
            "root_sec_style": self._GET_STRING(svm_info, "root-volume-security-style"),
            "root_name": self._GET_STRING(svm_info, "root-volume"),
            "root_retent": self._GET_STRING(svm_info, "volume-delete-retention-hours"),
            
            "locked": self._GET_BOOL(svm_info, "is-config-locked-for-changes"),
            "svm_state": self._GET_STRING(svm_info, "state"),
            
            "proto_enabled": self._GET_CONTENT_LIST(svm_info, "allowed-protocols"),
            "proto_inactive": self._GET_CONTENT_LIST(svm_info, "disallowed-protocols")
        }

        dd["oper_state"] = self._GET_STRING(svm_info, "operational-state")
        if self._GET_STRING(svm_info, "operational-state-stopped-reason"):
            dd["oper_state"] += " caused by: " + self._GET_STRING(
                        svm_info, "operational-state-stopped-reason")

        return {
            'success' : True, 'errmsg': '', 'data': dd}