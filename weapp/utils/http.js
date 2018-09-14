const startLoading = () => {
    wx.showLoading({
        title: '加载中',
    });
};

const stopLoading = () => {
    setTimeout(function () {
        wx.hideLoading();
    }, 300)
};

const request = (url, callback, loadingUI) => {
    if (!!loadingUI) {
        loadingUI.startLoading();
    } else {
        startLoading();
    }
    wx.request({
        url: url,
        header: {
            'content-type': 'application/json' // 默认值
        },
        success: function (res) {
            callback(res.data);
        },
        fail: function (error) {
            console.error(error);
        },
        complete: function () {
            if (!!loadingUI) {
                loadingUI.stopLoading();
            } else {
                stopLoading();
            }
        }
    })
};

const url = (uri) => {
  return encodeURI(getApp().globalData.service + uri);
};

module.exports = {
    request: request,
    url: url,
};
