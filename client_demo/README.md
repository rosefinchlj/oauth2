# GitHub OAuth 第三方登录示例教程



网站登录时，允许使用第三方网站的身份，这称为"第三方登录"。

![img](E:\use\note\pic\bg2019042101.jpg)





**下面就以 GitHub 为例，写一个最简单的应用，演示第三方登录**。

[github官方文档](https://docs.github.com/zh/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps)

## 一、第三方登录的原理

所谓第三方登录，实质就是 OAuth 授权。用户想要登录 A 网站，A 网站让用户提供第三方网站的数据，证明自己的身份。获取第三方网站的身份数据，就需要 OAuth 授权。

1. A 网站让用户跳转到 GitHub。
2. GitHub 要求用户登录，然后询问"A 网站要求获得 xx 权限，你是否同意？"
3. 用户同意，GitHub 就会重定向回 A 网站，同时发回一个授权码。
4. A 网站使用授权码，向 GitHub 请求令牌。
5. GitHub 返回令牌.
6. A 网站使用令牌，向 GitHub 请求用户数据。



## 二、应用登记

一个应用要求 OAuth 授权，必须先到对方网站登记，让对方知道是谁在请求。

访问这个[网址](https://github.com/settings/applications/new)，填写登记表。

![img](E:\use\note\pic\bg2019042102.jpg)



用的名称随便填，主页 URL 填写`http://localhost:8080`，跳转网址填写`http://localhost:8080/oauth/redirect`。

提交表单以后，GitHub 应该会返回客户端 ID（client ID）和客户端密钥（client secret），这就是应用的身份识别码。



## 三、示例仓库





## 四、浏览器跳转 GitHub



示例的首页很简单，就是一个链接，让用户跳转到 GitHub

![img](E:\use\note\pic\bg2019042103.jpg)

跳转的 URL 如下：

```
https://github.com/login/oauth/authorize?
  client_id=7e015d8ce32370079895&
  redirect_uri=http://localhost:8080/oauth/redirect
```

这个 URL 指向 GitHub 的 OAuth 授权网址，带有两个参数：`client_id`告诉 GitHub 谁在请求，`redirect_uri`是稍后跳转回来的网址。

用户点击到了 GitHub，GitHub 会要求用户登录，确保是本人在操作。



## 五、授权码

登录后，GitHub 询问用户，该应用正在请求数据，你是否同意授权。

![img](E:\use\note\pic\bg2019042104.png)

用户同意授权， GitHub 就会跳转到`redirect_uri`指定的跳转网址，并且带上授权码，跳转回来的 URL 就是下面的样子。

```
http://localhost:8080/oauth/redirect?
  code=859310e7cecc9196f4af
```



上面的地址为后端接口， 后端收到这个请求以后，就拿到了授权码（`code`参数）。



## 六、后端实现

``` go
http.HandleFunc("/oauth/redirect", handler)
```



**从url参数中拿到code(授权码)，为后去AccessToken使用**。



## 七、令牌

后端使用这个授权码，向 GitHub 请求令牌

```
POST https://github.com/login/oauth/access_token
```

| 名称            | 类型     | 说明                                                 |
| :-------------- | :------- | :--------------------------------------------------- |
| `client_id`     | `string` | **必填。** 从 GitHub 收到的 OAuth App 的客户端 ID。  |
| `client_secret` | `string` | **必填。** 从 GitHub 收到的 OAuth App 的客户端密码。 |
| `code`          | `string` | **必填。** 收到的作为对步骤 1 的响应的代码。         |
| `redirect_uri`  | `string` | 用户获得授权后被发送到的应用程序中的 URL。           |



默认情况下，响应采用以下形式：

```
access_token=gho_16C7e42F292c6912E7710c838347Ae178B4a&scope=repo%2Cgist&token_type=bearer
```

如果在 `Accept` 标头中提供格式，则还可以接收不同格式的响应。 例如 `Accept: application/json` 或 `Accept: application/xml`：

```json
Accept: application/json
{
  "access_token":"gho_16C7e42F292c6912E7710c838347Ae178B4a",
  "scope":"repo,gist",
  "token_type":"bearer"
}
```

``` xml
Accept: application/xml
<OAuth>
  <token_type>bearer</token_type>
  <scope>repo,gist</scope>
  <access_token>gho_16C7e42F292c6912E7710c838347Ae178B4a</access_token>
</OAuth>
```



## 八、API 数据

访问令牌可用于代表用户向 API 提出请求。

```
Authorization: Bearer OAUTH-TOKEN
GET https://api.github.com/user
```

例如，您可以像以下这样在 curl 中设置“授权”标头：

```shell
curl -H "Authorization: Bearer OAUTH-TOKEN" https://api.github.com/user
```



## 九、 说明

首先：

验证码是怎么来的呢，https://github.com/login/oauth/authorize?client_id=[client_id]&redirect_uri=[redirect_uri]这个链接发起的请求来的，client_id相当于是暴露在前端，通过浏览器的开发者工具就可以看到的，所以[任何人]都可以拿着这个client_id去向github发起请求，

然后：

[任何人]把redirect_uri换成自己的服务器的uri，那照你这样做，[任何人]都可以冒充客户端去拿到accessToken。
所以，之所以要多加验证码code这一步，就是要验证是否是client_id这个id所标识的客户端所发起的认证授权请求，
通过**网站后台**第二次的 code + client_secret 就可以确定是client_id这个id所标识的客户端发起的请求，因为就算别人冒充了你的client_id和伪造了redirect_uri，拿到了code,但是他没有client_secret，他也没办法拿到access_token。