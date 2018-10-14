'use strict';

function showAlert(text, type) {
  // toggle classes for appropiate styling ('success', 'error')
  $('#info').removeClass('success error');
  $('#info').addClass(type);

  $('html, body').animate({ scrollTop: 0 }, 400);
  $('#info').html(text).fadeIn(200).delay(5000).fadeOut(400);
}

$(document).ready(function() {
  // submit handler for forms
  $('form').submit(function() {
    var that = this;

    // submit via ajax
    $.ajax({
      data: $(that).serialize(),
      type: $(that).attr('method'),
      url:  $(that).attr('action'),
      beforeSend: function (xhr) {
        /* Authorization header */
        xhr.setRequestHeader("Authorization", localStorage.getItem('token'));
      },
      error: function(xhr, status, err) {
        showAlert(xhr.responseText, 'error');
      },
      success: function(res) {
        showAlert("Success doing what you tried to do", 'success');
      }
    });
    return false;
  });

  $('#form-login').submit(function() {
        var that = this;
        // submit via ajax
        $.ajax({
            data: $(that).serialize(),
            type: $(that).attr('method'),
            url:  $(that).attr('action'),

            error: function(xhr, status, err) {
                showAlert(xhr.responseText, 'error');
            },
            success: function(res) {
                console.log(res);
                localStorage.setItem('token', res); // write
                showAlert("Successfully authorized!", 'success');
            }
        });
        return false;
    });

    $('#form-viewUsers').submit(function() {
        var that = this;
        // submit via ajax
        $.ajax({
            data: $(that).serialize(),
            type: $(that).attr('method'),
            url:  $(that).attr('action'),
            beforeSend: function (xhr) {
                /* Authorization header */
                xhr.setRequestHeader("Authorization", localStorage.getItem('token'));
            },

            error: function(xhr, status, err) {
                showAlert(xhr.responseText, 'error');
            },
            success: function(res) {
                $("#userList").html("<h1>"+res+"</h1>");
                console.log(res);
            }
        });
        return false;
    });
});
