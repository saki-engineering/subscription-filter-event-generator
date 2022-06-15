# subscription-filter-event-generator

**EXAMPLES**
```bash
$ subscription-filter-event-generator  \
	--owner=000000000000 \
	--log-group=hsaki \
	--log-stream=mystream \
	--subscription-filter=myfilter \
	--log-events=file://example.json

{
	"awslogs": {
		"data": "H4sIAAAAAAAC/4TOQUvDQBAF4L8S3nkPSbqg7i1gWjyIkORWgsRmbBa72bAzVULpf5e2FFQSu6d985iPOcARc7OlahwIBo9Zlb0+52WZrXIo+K+eAgziHw8KO79dBb8fYNBx82Evo1ICNQ4GbuTLV4H3b7wJdhDr+6XdCQWGWcON7+eA+ryZf1Ivp+IA28KATvmpTaAg1hFL4waYROtY6/Th/i6OY3U9HAbrvCheijpa2sASCbFE1/KofpPpPJlMkSVtfN/+by7mzXTKrDobbpB6nlxMkVq6P2B9/A4AAP//BC+CX9sBAAA="
	}
}
```

```bash
$ subscription-filter-event-generator  \
	--owner=000000000000 \
	--log-group=hsaki \
	--log-stream=mystream \
	--subscription-filter=myfilter \
	--log-events='[{"id":"eventId1","message":"[ERROR] First test message","timestamp":1440442987000},{"id":"eventId2","message":"[ERROR] Second test message","timestamp":1440442987001}]'

{
	"awslogs": {
		"data": "H4sIAAAAAAAC/3SOT0vDQBDFv0p45z1sSkCdW8C0eBAh6a0Eic1YF7t/2JkqofS7S1sCKjqnefNjfrwjPIsMO15PiUG4r9f182PTdfWqgUH8DJxBsN8GBvu4W+V4SCC8yfDurqdOMw8eBD/JdTWQw4tss0vqYli6vXIW0AZ+er0E9JfP5oODnsERbgSBz/lhLGGgzrPo4BOorCpbVYu72xtrrZmLg7Bp2vap7Yuly6KFsmgxw5P5qVz8ryz/Una8jWH85exPXwEAAP//rEsFZzcBAAA="
	}
}
```

## Available Flags

| Option | Description |
|:------------|-------------|
| **`--owner`**    | ログ送信元が属するAWSアカウント (default: 123456789123) |
| **`--log-group`** | 送信元となるロググループ名 (default: "testLogGroup") |
| **`--log-stream`** | 送信元となるログストリーム名 (default: "testLogStream") |
| **`--subscription-filter`** | 送信元のサブスクリプションフィルタ名 (default: "testFilter") |
| **`--log-events`** | 送信するログイベントデータ (default: `sam local generate-event cloudwatch logs`コマンドで生成されるものと同じ) |

```bash
$ sam local generate-event cloudwatch logs | jq -r .awslogs.data | base64 -d | gzip -d
{
  // (略)
  "logEvents": [
    {
      "id": "eventId1",
      "timestamp": 1440442987000,
      "message": "[ERROR] First test message"
    },
    {
      "id": "eventId2",
      "timestamp": 1440442987001,
      "message": "[ERROR] Second test message"
    }
  ]
}
```
