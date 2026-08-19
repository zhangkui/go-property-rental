用户 `user-1` 修改密码后，新密码登录不了，旧密码却还能用，回归报 `stored hash does not accept new password`。输入错误的当前密码时也不能更新任何东西。把改密码流程修好，会话撤销也要针对当前用户。
