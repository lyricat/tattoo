var converter = null;
//var md_file = null;
var FORCE = 1;
var MODE_FULL = 1;
var html_begin = '<!DOCTYPE html>\n\
<html dir="ltr">\n\
    <meta charset="utf-8" />\n\
    <head>\n\
    <title>{TITLE}</title>\n\
    <style>\n\
{STYLE}\n\
#preview { margin: 0 auto; width: 800px;}\n\
    </style>\n\
    </head>\n\
    <body>\n\
        <div id="preview">';

var html_end = '</div></body></html>';

function update(mode) {
    if (!mode) mode = 0;
    if (can_update || mode === FORCE) {
        var source = $.trim(editor.getSession().getValue());
        if (source.lenght != 0) {
            var html = converter.makeHtml(source);
            $('#preview').html(html);
        }
    }
    setTimeout(update, 1000);
}
function onresize() {
    var view_h = $(this).height(); 
    var view_w = $(this).width(); 
    $('#container').height(view_h - $('#bar').height() - 1);
    $('#container').children('.pane')
        .height(view_h - $('#bar').height - 5);
    $('#input_pane').width(parseInt(view_w/2)+10)
    $('#meta_pane').width(parseInt(view_w/2)+10)
    $('#preview_pane').width(parseInt(view_w/2)-20);
}
function change_theme(theme) {
    if (theme == 'dark') {
        $('.ace_scroller, .ace_sb, .ace_editor').addClass('dark');
        editor.setTheme("ace/theme/twilight");
    } else {
        $('.ace_scroller, .ace_sb, .ace_editor').removeClass('dark');
        editor.setTheme("ace/theme/textmate");
    }
}

function save_state() {
    var state = [{name: '', data: editor.getSession().getValue()}];
    try {
        localStorage.setItem('files', JSON.stringify(state));
        //console.log('save state', JSON.stringify(state));
    } catch (e) {
        if (e == QUOTA_EXCEEDED_ERR) {
            alert('Quota exceeded!'); 
        }
    }
}

function resume_state() {
    var state = localStorage.getItem('files');
    if (state) {
        state = JSON.parse(state);
        //console.log('resume state', JSON.stringify(state));
        editor.getSession().setValue(state[0].data);
    }
}

function fake_click(obj) {
    var ev = document.createEvent("MouseEvents");
    ev.initMouseEvent(
        "click", true, false, window, 0, 0, 0, 0, 0
        , false, false, false, false, 0, null
        );
    obj.dispatchEvent(ev);
}
function popup_import() {
    var file = document.createElementNS("http://www.w3.org/1999/xhtml", "input");
    fake_click(file);
}
function export_raw(name, data) {
    if (!window.BlobBuilder) {
        window.BlobBuilder = window.WebKitBlobBuilder;
    }
    var urlObject = window.URL || window.webkitURL || window;
    var builder = new BlobBuilder(); 
    builder.append(data); 
    var export_blob = builder.getBlob(); 

    var save_link = document.createElementNS("http://www.w3.org/1999/xhtml", "a")
    save_link.href = urlObject.createObjectURL(export_blob);
    save_link.download = name;
    fake_click(save_link);
}

function export_html(mode) {
    var html = $('#preview').html();
    var all_text = $('#preview').text().split('\n');

    var name = 'Untitled';
    if (all_text.length != 0 && $.trim(all_text[0]).length != 0) {
        name = $.trim(all_text[0]);
    }
    var filename = name + '.html';
    if (mode == MODE_FULL) {
        export_raw(filename, html_begin.replace('{TITLE}', name) + html + html_end);
    } else {
        export_raw(filename, html);
    }
}

function export_source() {
    var source = editor.getSession().getValue(); 
    var all_text = $('#preview').text().split('\n');

    var name = 'Untitled';
    if (all_text.length != 0 && $.trim(all_text[0]).length != 0) {
        name = $.trim(all_text[0]);
    }
    var filename = name + '.md';
    export_raw(filename, source);
}

function load_source(file) {
    var reader = new FileReader();
    reader.onload = function (e) {
        editor.getSession().setValue(e.target.result);
    }
    reader.readAsText(file);
}

var editor = null;
var can_update = true;
$(document).ready(function () {
    $(window).resize(function (event) {
        onresize();
    });
    editor = ace.edit("input_pane");
    editor.getSession().setValue("the new text here");
    editor.getSession().setTabSize(4);
    editor.getSession().setUseSoftTabs(true);
    document.getElementById('input_pane').style.fontSize='14px';    
    editor.getSession().setUseWrapMode(true);
    editor.setShowPrintMargin(true);    
    var mode = require("ace/mode/markdown").Mode;
    editor.getSession().setMode(new mode());

    $('body').bind('dragover', function () {
        return false;    
    }).bind('dragend', function () {
        return false;
    }).bind('drop', function (ev) {
        var md_file = ev.originalEvent.dataTransfer.files[0]; 
        load_source(md_file);
        return false;
    });
    
    $('#import_file_button').hover(function () {
        $('#import_button').addClass('hover');
    }, function () {
        $('#import_button').removeClass('hover');
        $('#import_button').removeClass('active');
    }).mousedown(function () {
        $('#import_button').addClass('active');
    }).mouseup(function () {
        $('#import_button').removeClass('active');
    }).change(function () {
        load_source($(this).get(0).files[0]);
    });

    $('#export_html_button').click(function () {
        export_html();
        return false;
    });

    $('#export_full_html_button').click(function () {
        export_html(MODE_FULL);
        return false;
    });

    $('#export_source_button').click(function () {
        export_source();
        return false;
    });

    $('#export_button_wrapper').hover(function () {
    }, function () {
        $('#export_menu').slideUp();
    });
    $('#export_button').mousedown(function () {
        $('#export_menu').slideDown();
    });
    $('#save_button').click(function () {
        $('#text_box').val(editor.getSession().getValue())
        document.edit_form.submit();
    });
    $('#save_button').click(function () {
        $('#text_box').val(editor.getSession().getValue())
        document.edit_form.submit();
    });

    $('#color_scheme > a').click(function () {
        $('#color_scheme > a').removeClass('selected');
        $(this).addClass('selected');
        change_theme($(this).attr('href'));
        return false;
    })
    
    $('#preview_pane').hover(function () {
        can_update = false;
    }, function () {
        can_update = true;
    });

    $('#title_box').blur(function () {
        if ($('#url_box').val() == '') {
            var url = encodeURIComponent($(this).val().toLowerCase().replace(/\s/g, '-')); 
            $('#url_box').val(url)
        }
    });

    // load style for exporting
    $.get('/static/css/preview.css', function (data) {
        html_begin = html_begin.replace('{STYLE}', data);
    }); 

    converter = new Markdown.Converter();
    update(FORCE);
    onresize();
    setTimeout(function () {
        change_theme('dark');
        onresize();
        if ($.trim($('#text_box').val()).length == 0) {
            resume_state();
        } else {
            editor.getSession().setValue($('#text_box').val());
        }
    }, 10);
});

window.onunload = function () {
    save_state();
}

