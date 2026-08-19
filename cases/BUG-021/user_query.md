用户被停用以后，旧 token 还能继续用。把 `target-1` 设为 disabled 后，原来的 access token 仍能鉴权，refresh token 也能换新会话，而且被清掉的会话有时不是这个用户的。需要把停用后的登录状态处理好，正常 active 用户不要受影响。
