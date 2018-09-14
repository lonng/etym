// pages/improve/improve.js
const http = require('../../utils/http.js');

Page({

    /**
     * 页面的初始数据
     */
    data: {
        translation: null,
        word: null,
        original: "",
        prevTrans: "",
        nextTrans: ""
    },

    /**
     * 生命周期函数--监听页面加载
     */
    onLoad: function (options) {
        var improveParam = wx.getStorageSync('__improve');
        if (!improveParam || !improveParam.etym || !improveParam.etymCN) {
            wx.showToast({
                title: '请求错误',
                icon: 'none',
                duration: 2000
            });
            wx.navigateBack();
            return;
        }

        console.log(improveParam);

        this.setData({word: improveParam.word});
        this.setSentences(improveParam);
    },

    getBLen: function (str) {
        if (str == null) return 0;
        if (typeof str != "string") {
            str += "";
        }
        return str.replace(/[^\x00-\xff]/g, "01").length;
    },

    setSentences: function (result) {
        var wordCount = this.getBLen(result.etymCN) / 2;
        var height = Math.ceil(wordCount / 21) * 50 + 'rpx;';

        this.setData({
            word: result.word,
            translation: result.trans,
            original: result.etym,
            nextTrans: result.etymCN,
            prevTrans: (' ' + result.etymCN).slice(1), // 深拷贝
            height: height,
        })
    },

    changeTrans: function (event) {
        this.setData({
            nextTrans: event.detail.value,
        })
    },

    resetTrans: function () {
        this.setData({
            nextTrans: (' ' + this.data.prevTrans).slice(1),
        })
    },

    submitImprove: function () {
        var self = this;
        var dirty = this.data.prevTrans !== this.data.nextTrans;

        if (!dirty) {
            wx.showToast({
                title: '没有修改内容',
                icon: 'success',
                duration: 1000
            });
            return;
        }

        wx.showModal({
            title: '',
            content: '是否提交到服务器',
            success: function (res) {
                if (res.confirm) {
                    console.log('用户点击确定');
                    let payload = {
                        word: self.data.word,
                        translation: self.data.nextTrans,
                    };

                    wx.showLoading({title: "正在提交"});
                    // 提交到服务器
                    wx.request({
                        url: http.url('/review/improve'),
                        method: 'POST',
                        data: JSON.stringify(payload),
                        success: function () {
                            wx.showToast({
                                title: '提交成功',
                                icon: 'success',
                                duration: 1000
                            });
                            console.log("submit success")
                        },
                        fail: function (err) {
                            wx.showToast({
                                title: '提交失败',
                                icon: 'none',
                                duration: 1000
                            });
                            console.log("error",err)
                        },
                        complete: function () {
                            wx.hideLoading();
                            wx.navigateBack();
                        }
                    })

                } else if (res.cancel) {
                    console.log('用户点击取消')
                }
            }
        })
    },

    /**
     * 生命周期函数--监听页面初次渲染完成
     */
    onReady: function () {

    },

    /**
     * 生命周期函数--监听页面显示
     */
    onShow: function () {

    },

    /**
     * 生命周期函数--监听页面隐藏
     */
    onHide: function () {

    },

    /**
     * 生命周期函数--监听页面卸载
     */
    onUnload: function () {

    },

    /**
     * 页面相关事件处理函数--监听用户下拉动作
     */
    onPullDownRefresh: function () {

    },

    /**
     * 页面上拉触底事件的处理函数
     */
    onReachBottom: function () {

    },

    /**
     * 用户点击右上角分享
     */
    onShareAppMessage: function () {

    }
})