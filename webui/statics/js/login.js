var encrypt = new JSEncrypt();
encrypt.setPublicKey('MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDH0gU8EzVkoDcKg7Eewx1OWZS5d7fc1+1D5hLOlcvZvEv2DdYXPJi9/zyE4ZptSMAFy++69h/AryCRtDRyVH6SLAlE8pTgiz7pLSgqFf54O64PKPmFmF/sC/81VV7UwdrOatPy8iuoeutiN7V3wLFb/OpOTEdeUbq8ZITfKgmtSwIDAQAB');

var app = new Vue({
    el: "#login-panel",
    data: {
        username: "",
        password: "",
    },
    methods: {
        login: function () {
            var encrypting = {
                password: this.password,
                encrypted_at: Math.round(new Date().getTime() / 1000)
            };
            var encrypted = encrypt.encrypt(JSON.stringify(encrypting));
            $.ajax({
                type: 'POST',
                url: '/api/v1/login',
                data: JSON.stringify({
                    username: this.username,
                    password: encrypted
                }),
                dataType: 'json',
                success: function (data) {
                    if (data.review) {
                        window.location ='/review?token='+data.token;
                    }else {
                        window.location ='/edit?token='+data.token;
                    }
                },
                error:function (xhr) {
                    var res = JSON.parse(xhr.response);
                    console.log(res)
                }
            });
        }
    }
});
