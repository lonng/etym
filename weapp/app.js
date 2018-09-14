const http = require('/utils/http.js');

//app.js
App({
    onLaunch: function () {
        // 展示本地存储能力
        var logs = wx.getStorageSync('logs') || []
        var self = this;
        logs.unshift(Date.now())
        wx.setStorageSync('logs', logs)

        // 获取配置信息
        wx.request({
            url: self.globalData.service + "/wordlist",
            success:function(res){
                console.log(res);
                wx.setStorageSync("__wordlist", res.data.items);
            },
        });

        // 检查缓存API兼容版本, 如果不兼容, 则清空
        var apiVersion = wx.getStorageSync('cache_api_version') || self.globalData.apiVersion;
        if (apiVersion !== self.globalData.apiVersion) {
            wx.setStorageSync('cached_etymology', {})
            console.log("不兼容的API版本, 清空本地缓存")
        }
        wx.setStorageSync('cache_api_version', self.globalData.apiVersion)
    },
    globalData: {
        apiVersion: "v1.1",
        service: 'https://etym.apps.qilecloud.com/api/v1',
        // service: 'http://127.0.0.1:8080/api/v1',
        cacheList: [],
    }
})