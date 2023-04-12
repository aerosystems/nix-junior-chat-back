# nix-junior-chat-back

🍕Golang Backend App based on [ECHO]( https://echo.labstack.com/) & [GORM](https://gorm.io/) frameworks.

🔋Run App with simple command **_make up_** (of course if you have preinstalled Docker).

About most popular command read here - **_make help_**

📚Read & Test with [Swagger Docs](http://localhost:8080/docs/index.html)

🎲Test or Develop with Postman Collection(just import **postman-collection.json** file)

📌All environment variables(**_.env.dev_** file with passwords & API keys) are intentionally left in the root directory for easy application startup.

ℹ️Variable APP_ENV has two possible values:
- APP_ENV=dev respons debug info level of error
- APP_ENV=prod respons just message about error [DEFAULT]
