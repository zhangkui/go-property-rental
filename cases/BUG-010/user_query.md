同一笔账单收款重试后，对账流水号会变。客户端两次都传 `external-ref`，结果返回的 reference 不一样，后面的分摊记录也有可能找不到本次 payment。这个 reference 是上游用来认同一笔业务的，帮我把重复提交这块修一下，测试先别动。

go test ./scripts/verify -count=1 -run '^TestBug010_BusinessRegression$'