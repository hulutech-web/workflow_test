import { http } from "@/plugins/axios";
import router from "@/plugins/router";
import XEUtils from "xe-utils";
const storage = useStorage();

export default () => {
  const indexProcs=(entry_id)=>{
    return http.request({
      url: `proc/${entry_id}`,
      method: "GET",
    });
  }
  // 方法
  const setPass = async (data) => {
    return await http.request({
      url: `pass`,
      method: "POST",
      data:data
    });
  };

  const setUnPass = async (data) => {
    return await http.request({
      url: `unpass`,
      method: "POST",
      data: data,
    });
  };

  const setRevoke = async (data) => {
    return await http.request({
      url: `revoke`,
      method: "POST",
      data: data,
    });
  };

  const addSign = async (data) => {
    return await http.request({
      url: `addsign`,
      method: "POST",
      data: data,
    });
  };

  const transferProc = async (data) => {
    return await http.request({
      url: `transfer`,
      method: "POST",
      data: data,
    });
  };

  const addComment = async (data) => {
    return await http.request({
      url: `comment`,
      method: "POST",
      data: data,
    });
  };

  const getComments = async (entry_id) => {
    return await http.request({
      url: `comments/${entry_id}`,
      method: "GET",
    });
  };

  return {
    setPass,
    setUnPass,
    setRevoke,
    addSign,
    transferProc,
    addComment,
    getComments,
    indexProcs
  };
};
