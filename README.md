KindlePush
===
Subscribe rss feeds and delivered to your Kindle.

Requirements
===
[KindleGen](https://www.amazon.com/gp/feature.html?docId=1000765211) : KindleGen converts source content to Mobi formats file,officially supported by Amazon.

How To Use
===
``` kindlepush --help```

command line options:

- max-number=50 : the maximum number of items to fetch.

- kindle=xx@kindle.com : your kindle email address.

- subscribes=engadget_en,techcrunch_en, ... : the id of rss feed to subscribed, use a comma(,) to separate.

- email-from=xx@xx : your email address.

- email-username=xx : your email account name.

- email-password=xx : your email account password.

- email-smtp=smtp.xx.xx:25 : the email SMTP address with port number.

**KindlePush use your email account to send eBook to your kindle device address on the client, it's safe.**

RSS feeds
===
|ID |Name |
|--------------------------|----------------|
|engadget_en | [Engadget](http://www.engadget.com/)|
|engadget_cn | [Engadget 中文版](http://cn.engadget.com/)|
|techcrunch_en | [Techcrunch](http://feeds.feedburner.com/)|
|techcrunch_cn | [Techcrunch 中文版](http://techcrunch.cn)|
|zhihu | [知乎每日精选](https://www.zhihu.com)|
|blog_msdn_dotnet| [MSDN dotnet blog](https://blogs.msdn.microsoft.com/dotnet)|

welcome to contribute to more plugin.

End