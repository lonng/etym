var app = new Vue({
    el: "#review-panel",
    data: {
        translationId: -1,
        backup: null,
        prevTrans: null,
        nextTrans: null,
        original: null,
        word: '',
    },

    methods: {
        submit: function (event) {
            var self = this;
            var data = JSON.stringify({
                translation_id: self.translationId,
                word: self.word,
                translation: self.nextTrans,
            });
            console.log(data);
            $.ajax({
                type: 'POST',
                url: '/api/v1/improve/edit?token='+(new URL(window.location.href).searchParams.get('token')),
                data: data,
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
        },
        finalSubmit: function () {
            var self = this;
            $.ajax({
                type: 'GET',
                url: '/api/v1/review/approve?id='+self.translationId+'&final=1',
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
        }
    }
});

function nextReview() {
    $.ajax({
        type: 'GET',
        url: '/api/v1/improve/edit?token='+(new URL(window.location.href).searchParams.get('token')),
        dataType: 'json',
        success: function (data) {
            console.log(data);
            app.translationId = data.translation_id;
            app.nextTrans = data.next;
            app.prevTrans = (' ' + data.next).slice(1); // 深拷贝
            app.original = data.original;
            app.word = data.word;
        },
        error:function (xhr) {
            var res = JSON.parse(xhr.response);
            console.log(res)
            if (res.error.search('permission denied') >= 0) {
                window.location = '/login'
            }
        }
    });
}

nextReview();