// pages/index/rand.index.js.js
const http = require('../../utils/http.js');
const app = getApp()

const selectedKey = "__selected";

Page({

    /**
     * 页面的初始数据
     */
    data: {
        isRandMode: false,
        currentEtym: null,
        searchText: "",
        scrollTop: 0,
        searchList: null,
        isSearching: false,
        isShowMenu: false,

        colcount: 3,
        menulist: null,

        _cacheSearchList: null
    },

    /**
     * 生命周期函数--监听页面加载
     */
    onLoad: function (options) {
        // 显示具体单词
        if (!!options.word) {
            return this.detail(options.word);
        }

        // 随机模式
        this.setData({isRandMode: true});

        // 设置菜单
        var wordlist = wx.getStorageSync('__wordlist');
        if (!!wordlist && wordlist.length > 0) {
            var seleted = wx.getStorageSync(selectedKey) || [];
            if (seleted.length < 1) {
                console.log("没有选择单词范围, 默认使用Coca20000");
                seleted = [];
                for (var i = 0; i < wordlist.length; i++) {
                    if (wordlist[i].desc.indexOf('COCA') >= 0) {
                        seleted.push(wordlist[i].value);
                        break;
                    }
                }
            }

            var colcount = this.data.colcount;
            var rowcount = Math.ceil(wordlist.length / colcount);
            var menulist = [];
            for (var row = 0; row < rowcount; row++) {
                var menurow = [];
                for (var col = 0; col < colcount; col++) {
                    var index = row * colcount + col;
                    if (index < wordlist.length) {
                        wordlist[index].checked = true;
                        menurow.push(wordlist[index])
                    } else {
                        menurow.push({})
                    }
                }
                menulist.push(menurow)
            }
            console.log("===>", wordlist);
            console.log("===>", menulist);
            this.setData({menulist: menulist});
            this.initMenulist(seleted)
        }

        // 显示缓存单词
        var cacheList = wx.getStorageSync('cache-history') || [];
        // unique word list
        if (!!cacheList && cacheList.length > 0) {
            console.log("use cache", cacheList[0]);
            this.detail(cacheList[0].word)
        } else {
            // 随机一个单词
            this.refresh();
        }
    },

    newline: function (trans) {
        if (!!trans.translation) {
            let trimed = trans.translation.trim();
            let res = trimed.split("\\n").join("\n");
            trans.translation = res;
        }
    },

    // 处理换行
    wordPipe: function (etym) {
        console.log(etym);
        if (!!etym.trans) {
            this.newline(etym.trans)
        }

        if (!!etym.related && etym.related.length > 0) {
            for (var i = 0; i < etym.related.length; i++) {
                this.newline(etym.related[i]);
            }
        }

        if (!!etym.ref && etym.ref.length > 0) {
            for (var i = 0; i < etym.ref.length; i++) {
                if (!!etym.ref[i].dict) {
                    this.newline(etym.ref[i].dict);
                }

                if (!!etym.ref[i].related && etym.ref[i].related.length > 0) {
                    for (var j = 0; j < etym.ref[i].related.length; j++) {
                        this.newline(etym.ref[i].related[j]);
                    }
                }
            }
        }
    },

    scrollToTop: function () {
        this.setData({
            scrollTop: 0,
        })
    },

    storageKey: function (word) {
        return word;
    },

    // 请求新的词源回掉
    setCurrentEtym: function (result) {
        if (!result) {
            return;
        }

        this.wordPipe(result);

        // TODO: 默认使用中文显示
        if (!!result.etym && result.etym.length > 0) {
            for (var i = 0; i < result.etym.length; i++) {
                result.etym[i].showCN = false;
            }
        }

        // replace with new etym list
        this.setData({
            currentEtym: result,
        });

        if (!!result.trans) {
            try {
                var cacheList = wx.getStorageSync('cache-history') || [];
                console.log(result);

                // unique word list
                if (!!cacheList && cacheList.length > 0) {
                    for (var i = 0; i < cacheList.length; i++) {
                        if (cacheList[i].word === result.trans.word) {
                            cacheList.splice(i, 1);
                        }
                    }
                }
                cacheList.push(result.trans);

                // 只保留最近10个
                if (cacheList.length > 10) {
                    var tmp = cacheList.slice(cacheList.length - 10);
                    cacheList = tmp;
                }

                // 保存数据
                try {
                    wx.setStorageSync('cache-history', cacheList)
                } catch (e) {
                    console.log("保存历史记录失败", e)
                }
            } catch (e) {
                console.log("更新历史记录失败", e);
            }
        }


        // 保存本地缓存
        var localCache = wx.getStorageSync('cached_etymology') || {}
        var wordKey = this.storageKey(result.word)
        if (!localCache[wordKey]) {
            console.log(result);
            localCache[wordKey] = result;
            wx.setStorageSync('cached_etymology', localCache);
        }

        this.scrollToTop();
    },

    // 刷新随机单词
    refresh: function () {
        var ranges = wx.getStorageSync(selectedKey) || [];
        http.request(getApp().globalData.service + '/rand?range=' + ranges.join(','), this.setCurrentEtym);
    },

    detail: function (word) {
        // 检查本地缓存
        var localCache = wx.getStorageSync('cached_etymology') || {};
        var wordKey = this.storageKey(word);
        if (!!localCache[wordKey]) {
            wx.showLoading({title: '加载中',});
            this.setCurrentEtym(localCache[wordKey]);
            setTimeout(wx.hideLoading, 300);
            return;
        }

        http.request(http.url('/etym?word=' + word), this.setCurrentEtym);
    },

    // 单词详情
    searchDetail: function (event) {
        if (!event.currentTarget.dataset.word) {
            return;
        }
        this.detail(event.currentTarget.dataset.word)
    },

    // 单词详情
    wordDetail: function (event) {
        if (!event.currentTarget.dataset.word) {
            return;
        }
        wx.navigateTo({
            url: "index?word=" + event.currentTarget.dataset.word,
        })
    },

    updateSearchPanel: function (result) {
        console.log("search result=>", result);
        if (!!result && result.length > 0) {
            for (var i = 0; i < result.length; i++) {
                if (!!result[i].translation) {
                    let trimed = result[i].translation.trim();
                    let res = trimed.split("\\n").join(" ");
                    result[i].translation = res;
                }
            }
        }
        this.setData({
            searchList: result
        })
    },

    showHistory: function () {
        try {
            var value = wx.getStorageSync('cache-history')
            if (value) {
                this.updateSearchPanel(value.reverse());
            }
        } catch (e) {
            console.log("获取历史记录失败", e);
        }
    },

    searchTextInput: function (event) {
        let word = event.detail.value.trimLeft();
        this.setData({searchText: word});
        if (!word || word === '') {
            this.showHistory();
            return word;
        }

        this.setData({searchList: null});

        let self = this;
        let loadingUI = {
            startLoading: function () {
                self.setData({isSearching: true})
            },
            stopLoading: function () {
                self.setData({isSearching: false})
            }
        };
        http.request(http.url('/search?word=' + word), this.updateSearchPanel, loadingUI);
        return word;
    },

    searchFocus: function (event) {
        // 如果输入框已经有内容, 显示缓存搜索结果
        // 如果输入框没有内容, 显示最近查看过的单词
        if (!this.data.searchText || this.data.searchText.trim() === '') {
            this.showHistory()
        } else {
            if (!!this.data._cacheSearchList && this.data._cacheSearchList.length > 0) {
                this.updateSearchPanel(this.data._cacheSearchList);
            }
        }
    },

    searchBlur: function (event) {
        if (!!this.data.searchList && this.data.searchList.length > 0) {
            this.setData({_cacheSearchList: this.data.searchList})
        }
        this.updateSearchPanel(null);
    },

    clearSearch: function () {
        console.log('HHHH');
        this.setData({searchText: ''});
    },

    switchMenu: function () {
        this.setData({
            isShowMenu: !this.data.isShowMenu,
        })
    },

    playAudio: function (event) {
        const innerAudioContext = wx.createInnerAudioContext()
        innerAudioContext.autoplay = true
        innerAudioContext.src = event.currentTarget.dataset.url;
        innerAudioContext.onPlay(() => {
            console.log('开始播放')
        });
        innerAudioContext.onError((res) => {
            console.log(res.errCode, res.errMsg)
        })
    },

    switchLang: function (event) {
        var etymIndex = event.currentTarget.dataset.etymindex;
        var currentEtym = this.data.currentEtym;
        currentEtym.etym[etymIndex].showCN = !currentEtym.etym[etymIndex].showCN;
        this.setData({
            currentEtym: currentEtym,
        })
    },

    impoveTransaltion: function (event) {
        console.log(event.currentTarget.dataset);
        if (event.currentTarget.dataset.final) {
            return;
        }

        if (!event.currentTarget.dataset.iscn) {
            return;
        }

        var dataset = event.currentTarget.dataset;
        var improveParam = {
            word: dataset.word,
            trans: dataset.trans,
            etym: dataset.etym,
            etymCN: dataset.etymcn,
        };

        console.log("impove=>", improveParam);
        wx.showLoading({
            title: '进入纠错页',
        });

        wx.setStorageSync('__improve', improveParam);
        wx.navigateTo({
            url: "/pages/improve/improve",
            complete: function () {
                setTimeout(function () {
                    wx.hideLoading();
                }, 500)
            }
        })
    },

    initMenulist(values) {
        var menulist = this.data.menulist;
        for (var row = 0, rows = menulist.length; row < rows; ++row) {
            var rowItems = menulist[row];
            for (var col = 0, cols = rowItems.length; col < cols; ++col) {
                menulist[row][col].checked = false;

                for (var j = 0, lenJ = values.length; j < lenJ; ++j) {
                    if (menulist[row][col].value == values[j]) {
                        menulist[row][col].checked = true;
                        break;
                    }
                }
            }
        }
        this.setData({
            menulist: menulist
        });

        wx.setStorageSync(selectedKey, values)
    },

    // 选择范围复选款
    checkboxChange: function (event) {
        console.log(event.detail);
        this.initMenulist(event.detail.value);
    },

    etymContainerFocus: function() {
        this.searchBlur()
        if (this.data.isShowMenu) {
            this.setData({isShowMenu: false})
        }
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
        //this.refresh();
    },

    /**
     * 页面上拉触底事件的处理函数
     */
    onReachBottom: function () {
        //this.refresh();
        console.log("onReachBottom")
    },

    /**
     * 用户点击右上角分享
     */
    onShareAppMessage: function () {
        if (!this.data.isRandMode) {
            return;
        }
        if (!this.data.currentEtym || !this.data.currentEtym.word) {
            return;
        }
        console.log("onShareAppMessage", this.data.currentEtym.word);
        return {
            title: '"' + this.data.currentEtym.word + '"的词源',
            path: "/pages/index/index?word=" + this.data.currentEtym.word,
        }
    },
});