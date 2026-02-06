# 翻棋军旗 · 通信协议设计（WebSocket · v1）backend_protocol.md

## 一、通用格式
```json
{
  "type": "消息类型",
  "data": {}
}
```
## 二、客户端 → 服务端
翻棋
```json
{
  "type": "flip",
  "data": { "x": 3, "y": 5 }
}
```
走棋
```json
{
  "type": "move",
  "data": {
    "fromX": 3,
    "fromY": 5,
    "toX": 3,
    "toY": 6
  }
}
```
心跳
```json
{ "type": "ping" }

## 三、服务端 → 客户端
游戏开始
{
  "type": "start",
  "data": {
    "youCamp": "red",
    "turn": "red"
  }
}
```
状态同步
```json
{
  "type": "sync",
  "data": {
    "board": [],
    "turn": "blue"
  }
}
```
吃子结果
```json
{
  "type": "battle",
  "data": {
    "from": [3,5],
    "to": [3,6],
    "attacker": "团长",
    "defender": "营长",
    "result": "attacker_win"
  }
}
```

非法操作
```json
{
  "type": "error",
  "data": {
    "msg": "not your turn"
  }
}
```

游戏结束
```json
{
  "type": "game_over",
  "data": {
    "winner": "red",
    "reason": "flag_captured"
  }
}
```

## 四、设计原则

服务端为裁判

客户端只提交操作

sync 为最终状态

所有状态以服务端为准

## 五、扩展字段
```json
{
  "step": 12,
  "lastMove": {}
}
```
