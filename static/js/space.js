var space = "";
var url = "";
$(function () {
    var str = window.location.pathname;
    space = str.substring(str.lastIndexOf("/"));
    url = "/directory" + space;
    toastr.options.positionClass = 'toast-bottom-right';
    $("#file-panel").hide();
    datasPanel(datas);
});
function datasPanel(msg) {
    var count = 0;
    var total = 0;
    var data = "";
    //清空
    $("#datas-panel").html(data);
    var prefix = "<a href='/spaces" + space + "/";
    len = msg.length;
    for (i = 0; i < len; i++) {
        data += "<tr><td><a href='/spaces" + space + "/" + msg[i].name + "'>" + msg[i].name + "</a></td>";
        var size = parseInt(msg[i].size);
        data += "<td>" + size.toLocaleString() + "</td>";
        data += "<td>" + String(msg[i].time).substring(2) + "</td>";
        data += "<td><input class='file-checkbox' type='checkbox' name='" + i + "' onclick='ShowChoice()'></td></tr>";
        count++;
        total = total + size;
    }
    //加载数据
    $("#datas-panel").html(data);
    $("#assemble").text(count + " 个文件 " + total.toLocaleString() + " KB");
}

function DelFile() {
    var filename = new Array();
    $(".file-checkbox").each(function () {
        if ($(this).is(':checked')) {
            filename.push($(this).parent().prev().prev().prev().text());
        }
    });
    //没有选择时退出
    if (filename.length < 1) return;
    var o = new Object();
    o.filename = filename;
    $.ajax({
        type: 'DELETE',
        url: url,
        data: JSON.stringify(o),
        contentType: 'application/json',
        cache: false,
        success: function (msg) {
            datasPanel(msg);
            toastr.success("删除文件完成");
        },
        error: function (msg) {
            datasPanel(msg.responseJSON);
            toastr.error("删除文件出错");
        },
        complete: ShowHome(),
    });
}

var lock = false;
function UpLoadFile() {
    if (lock) return;
    var str = $("#userfile").val();
    if (str.length == 0) {
        toastr.warning("请选择文件");
        return;
    }
    NProgress.start();
    // FormData 对象
    var form = new FormData(document.forms.namedItem("uploadForm"));
    // XMLHttpRequest 对象
    var xhr = new XMLHttpRequest();
    xhr.open("post", url, true);
    // xhr.upload 储存上传过程中的信息
    xhr.upload.onprogress = function (ev) {
        if (ev.lengthComputable) {
            NProgress.set(ev.loaded / ev.total);
            $("#upload").text(Math.trunc(ev.loaded / 1024).toLocaleString() + "/" + Math.trunc(ev.total / 1024)
                .toLocaleString());
        }
    };
    xhr.onload = function (oEvent) {
        if (xhr.status == 200) {
            datasPanel(JSON.parse(xhr.response));
            $("#userfile").val("");
            $("#upload").text("上传");
            toastr.success("上传完成");
            setTimeout(function () {
                NProgress.done();
            }, 1000);
        }
        if (xhr.status == 500) {
            datasPanel(JSON.parse(xhr.response));
            $("#userfile").val("");
            $("#upload").text("上传");
            toastr.error("上传失败");
            setTimeout(function () {
                NProgress.done();
            }, 1000);
        }
    };
    xhr.onerror=function(oEvent){
        $("#userfile").val("");
        $("#upload").text("上传");
        toastr.error("上传未完成");
            setTimeout(function () {
                NProgress.done();
            }, 1000);
    };
    xhr.onreadystatechange = function () {
        switch (xhr.readyState) {
            case 4:
                lock = false;
                break;
        }
    };
    lock = true;
    //发送数据
    xhr.send(form);
}

function ChoiceAll() {
    ShowChoice();
    if ($("#check-all").is(':checked')) {
        $(".file-checkbox").prop("checked", true);
    } else {
        $(".file-checkbox").prop("checked", false);
        ShowHome();
    }
}

function ShowChoice() {
    $("#upload-panel").hide();
    $("#file-panel").show();
}

function ShowHome() {
    $(".file-checkbox").prop("checked", false)
    $("#file-panel").hide();
    $("#upload-panel").show();
}
