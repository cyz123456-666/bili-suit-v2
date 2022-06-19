# 简介

**傻瓜式，扫码登录**

**尽量模拟了抢购过程（风控 -1）**

**请遵守 GPLv3 开源协议！！！**

## 食用方法
1. 下载解压 [Release](https://github.com/MikaNyaru/bili-suit-v2/releases) 中对应的文件，哪个平台就用哪个
2. 填写 `config.json` 中的 `item_id` （装扮ID）
3. 运行脚本: 在终端中运行 `./bili-suit-tool`, Windows 用户请运行 ./bili-suit-tool.exe

## 小提示
1. 使用 `-c` 可指定配置文件， 示例: `./bili-suit-tool -c /etc/bili/1.json`
2. `cookies` 必要参数留空可使用扫码登录
3. `bp_enough` 为 true 时开启 b币余额校验，b币余额不足时不下单，为 false 将会直接下单

## 配置文件

**config.json**

```
{
  "bp_enough": true,
  "buy_num": "1",
  "coupon_token": "",
  "device": "android",
  "item_id": "",
  "time_before": 50,
  "cookies": {
    "SESSDATA": "",
    "bili_jct": "",
    "DedeUserID": "",
    "DedeUserID__ckMd5": ""
  }
}
```

# Author
[**超急玛丽**](https://space.bilibili.com/24924450)  
[**恋利普贝当**](https://space.bilibili.com/2932835)

