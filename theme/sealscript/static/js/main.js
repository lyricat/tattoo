function loadpage(hash) {
    if (hash.indexOf('#pos=') == 0) { // articles
        _loadpage('/?' + hash.substring(1));
    } else if (hash.indexOf('#t/') == 0) { // tag
        _loadpage('/' + hash.substring(1));
    } else if (hash.indexOf('#comment_') == 0) { // comment
    
    } else { // single
        _loadpage('/' + hash.substring(1));
    }
}

function _loadpage (url, callback) {
    $.ajax({
        url: url,
    success: function (data) {
        xmlDoc = $(data);
        var content_node = null;
        var next_node = null;
        var prev_node = null;
        for (var i = 0; i < xmlDoc.length; i+= 1) {
            var node = xmlDoc[i];
            if (node.nodeType == 1) {
                switch (node.getAttribute('id')) {
                case 'content': 
                    content_node = node;
                    break;
                case 'next':
                    next_node = node;
                    break;
                case 'prev':
                    prev_node = node;
                    break;
                default: break; 
                }
            }
        }
        if (next_node != null
            && next_node.style.display == 'block') {
            $('#next').attr('href', next_node.href);
            $('#next').attr('hash', next_node.hash);
            $('#next').show();
        } else {
            $('#next').hide();
        }
        if (prev_node != null
            && prev_node.style.display == 'block') {
            $('#prev').attr('href', prev_node.href);
            $('#prev').attr('hash', prev_node.hash);
            $('#prev').show();
        } else {
            $('#prev').hide();
        }
            
        if (content_node != null) {
            $('#content').empty();
            $('#content').html(content_node.innerHTML)
        }
        if (next_node != null) {
            window.location.hash = next_node.getAttribute('hash')
        }
        if (callback != undefined) {
            callback(data)
        }
    },
    error: callback
    });
}
var dynamic = false;
$(document).ready(function () {
    $('#recent_comments li').hover(function () {
        $(this).children('.bubble').show();
    }, function () {
        $(this).children('.bubble').hide();
    });
    if (dynamic) {
        var hash = window.location.hash;
        if (hash.length != 0) {
            loadpage(hash);
        }
    }

    $('.v_nav').click(function () {
        var icon = $(this).children('.icon');
        icon.addClass('loading');
        if (dynamic) {
            var url = $(this).attr('href');
            window.location.hash = $(this).attr('hash');
            _loadpage(url, function () {
                icon.removeClass('loading');
            });
            return false;
        }
    })
});
