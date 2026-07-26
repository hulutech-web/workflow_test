import { http } from "@/plugins/axios";
import XEUtils from "xe-utils";

export default () => {
  const showArchive = async (id: number) => {
    return await http.request({
      url: `archive/${id}`,
      method: "GET",
    });
  };

  const gridOptions = reactive<VxeGridProps<RowVO>>({
    border: "full",
    size: "small",
    showHeaderOverflow: true,
    showOverflow: true,
    keepSource: true,
    autoResize: true,

    formConfig: {
      titleWidth: 80,
      titleAlign: "right",
      items: [
        {
          field: "title",
          title: "标题",
          span: 12,
          titlePrefix: {
            useHTML: true,
            message: "模糊查询",
            icon: "vxe-icon-question-circle-fill",
          },
          itemRender: { name: "$input", props: { placeholder: "请输入标题" } },
        },
        {
          field: "status",
          title: "状态",
          span: 12,
          itemRender: {
            name: "$select",
            props: {
              placeholder: "请选择状态",
              options: [
                { value: 9, label: "已完成" },
                { value: -1, label: "已驳回" },
                { value: -2, label: "已撤回" },
              ],
            },
          },
        },
        {
          span: 24,
          align: "left",
          collapseNode: true,
          itemRender: {
            name: "$buttons",
            children: [
              {
                props: {
                  type: "submit",
                  content: "搜索",
                  status: "primary",
                },
              },
              { props: { type: "reset", content: "重置" } },
            ],
          },
        },
      ],
    },
    stripe: true,
    id: "archive_grid",
    rowConfig: {
      keyField: "id",
      isHover: true,
      useKey: true,
    },
    columnConfig: {
      resizable: true,
    },
    sortConfig: {
      trigger: "cell",
      remote: true,
    },
    filterConfig: {
      remote: true,
    },
    pagerConfig: {
      enabled: true,
      pageSize: 10,
      pageSizes: [5, 10, 15, 20, 50, 100],
    },

    toolbarConfig: {
      buttons: [],
      refresh: true,
      import: false,
      export: true,
      print: true,
      zoom: true,
      custom: true,
    },
    proxyConfig: {
      seq: true,
      sort: true,
      filter: true,
      form: true,
      props: {
        result: "data.data",
        total: "data.total",
      },
      ajax: {
        query: ({ page, sorts, filters, form }) => {
          const queryParams: any = Object.assign({}, form);
          const firstSort = sorts[0];
          if (firstSort) {
            queryParams.sort = firstSort.field;
            queryParams.order = firstSort.order;
          }
          filters.forEach(({ field, values }) => {
            queryParams[field] = values.join(",");
          });

          return http.request({
            url: `archive?pageSize=${page.pageSize}&currentPage=${
              page.currentPage
            }&${XEUtils.serialize(queryParams)}`,
            method: "GET",
          });
        },
      },
    },
    columns: [
      { field: "id", title: "ID", width: 60 },
      { field: "title", title: "标题", minWidth: 180, slots: { default: "title_default" } },
      { field: "status", title: "状态", width: 100, slots: { default: "status" } },
      { field: "created_at", title: "归档时间", width: 160 },
      { title: "操作", width: 120, slots: { default: "action" } },
    ],
  });

  return {
    showArchive,
    gridOptions,
  };
};
