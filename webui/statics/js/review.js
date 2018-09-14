var app = new Vue({
    el: "#review-panel",
    data: {
        translationId: -1,
        original: "",
        prevTrans: "",
        nextTrans: "",
        word: "",
        pendingReview: -1,
    },

    methods: {
        approve: function(final) {
            var self = this;
            $.ajax({
                type: 'GET',
                url: '/api/v1/review/approve?id='+self.translationId+'&final='+final,
                headers: {
                    'Authorization': (new URL(window.location.href).searchParams.get('token')),
                },
                success: function (data) {
                    console.log(data);
                    window.location = window.location
                },
                error:function (xhr) {
                    var res = JSON.parse(xhr.response);
                    console.log(res);
                }
            });
        },
        submit: function (event) {
            this.approve(0);
        },
        finalSubmit: function (event) {
            this.approve(1);
        },
        reject: function (event) {
            var self = this;
            $.ajax({
                type: 'POST',
                url: '/api/v1/improve/reject?token='+(new URL(window.location.href).searchParams.get('token')),
                data: JSON.stringify({
                    translation_id: self.translationId,
                }),
                processData: false,
                contentType: 'application/json; charset=utf-8',
                success: function (data) {
                    console.log(data);
                    window.location = window.location
                },
                error:function (xhr) {
                    var res = JSON.parse(xhr.response);
                    console.log(res)
                }
            });
        }
    }
});

function nextReview() {
    $.ajax({
        type: 'GET',
        url: '/api/v1/improve/review?token='+(new URL(window.location.href).searchParams.get('token')),
        dataType: 'json',
        success: function (res) {
            app.pendingReview = res.total;
            if (res.total > 0) {
                var historyInfo = res.history_info;
                console.log(res);
                app.translationId = historyInfo.translation_id;
                app.nextTrans = historyInfo.next;
                app.prevTrans = historyInfo.prev; // 深拷贝
                app.original = historyInfo.original;
                app.word = historyInfo.word;
            }
        },
        error:function (xhr) {
            var res = JSON.parse(xhr.response);
            console.log(res)
        }
    });
}

nextReview();