package main

import (
	"google.golang.org/protobuf/compiler/protogen"
)

func generateConfigCenter(g *protogen.GeneratedFile) {
	g.P("func GetConfig(ctx context.Context, key string) (*", sideCarGenPkg.Ident("GetConfigResponse"), ", error) {")
	g.P(`resp := &`, sideCarGenPkg.Ident("GetConfigResponse{}"), `
err := client.UnaryRPCRequest(&scrpc.RequestContext{
Ctx: ctx,
Req: &`, sideCarGenPkg.Ident("GetConfigReq{Key: key}"), `,
MessageType: `, scrpcGenPkg.Ident("Header_CONFIG_CENTER.Enum()"), `,
SenderService: "`, defaultCfg.Service, `",
Resp: resp,
})`)
	g.P("return resp, err")
	g.P("}")
}
