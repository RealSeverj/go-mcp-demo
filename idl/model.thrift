namespace go model
include "openapi.thrift"
struct BaseResp {
    1: i64 code (api.body="code", openapi.property='{
        title: "状态码",
        description: "响应状态码",
        type: "integer"
    }')
    2: string msg (api.body="msg", openapi.property='{
        title: "消息",
        description: "响应消息",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "基础响应",
        description: "所有响应的基础结构",
        required: ["code", "msg"]
    }'
)