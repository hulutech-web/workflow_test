import { http } from "@/plugins/axios";
import router from "@/plugins/router";
import XEUtils from "xe-utils";
const storage = useStorage();

export default () => {
  // 方法
  const loadFlowEntryConfig = async (id) => {
    return await http.request({
      url: `flow/${id}/entry`,
      method: "GET",
    });
  };

  const storeEntry = async (data) => {
    return await http.request({
      url: `entry`,
      method: "POST",
      data: data,
    });
  };
  const updateEntry = async (id, data) => {
    return await http.request({
      url: `entry/${id}`,
      method: "PUT",
      data: data,
    });
  };
  const showEntry = async (id) => {
    return await http.request({
      url: `entry/${id}`,
      method: "GET",
    });
  };
  const getEntryData = async (id) => {
    return await http.request({
      url: `entry/${id}/entrydata`,
      method: "GET",
    });
  };

  const resendEntry=(id)=>{
    return http.request({
      url:`entry/resend`,
      method:'POST',
      data:{entry_id:id}
    })
  }

  const revokeEntry=(id)=>{
    return http.request({
      url:`entry/revoke`,
      method:'POST',
      data:{entry_id:id}
    })
  }

  return {
    loadFlowEntryConfig,
    showEntry,
    storeEntry,
    updateEntry,
    getEntryData,
    resendEntry,
    revokeEntry
  };
};
