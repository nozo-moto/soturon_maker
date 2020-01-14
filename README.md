# Soturon_Maker

## What
某大学の卒論をMarkdownでかけるようにするためのCLIツール

## require

- pandoc
    ``` sh
    brew install pandoc
    ```
- 某大学への公開鍵認証
    ssh2を使う必要があるので、大学で鍵つくって落としてくるとかすると良い

## Install
``` sh
go get -u github.com/nozo-moto/soturon_maker
```

# How to RUN
1. write .env like .env_sample
2. run command

``` sh
cp .env_sample .env
soturon_maker run --file Theis.md
```

# 参考
偉大なる先輩の [記事](http://mizukisonoko.hatenablog.com/entry/2017/03/09/123213) を参考にしました

