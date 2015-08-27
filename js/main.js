(function() {
  'use strict';

  // Function hoisting = dumb
  /**
   * Build DOM??
   */
  var dom = {
    h1: function() {
      // Do whatever
      console.log('me are h1');
      return this; // Please don't ruin my code, `this`
    },
    h2: function() {
      console.log('me are h2');
      return this;
    }
  }

  ajax({
    url: 'http://127.0.0.1:3000/',
    method: 'GET',
    data: {
      url: '',
      keywords: '',
      email: '',
    },
    success: function() {
      console.log('Success');
    },
    error: function() {
      console.log('Error');
    }
  });

  // 'Dreamcode'
  dom.h1().h2();
})();

/**
 * Does an AJAX request.
 *
 * @param {Object} the setings
 */
function ajax(req) {

  var xhr = new XMLHttpRequest();
  xhr.open(req.method, req.url);

  // Be nice and set everything properly
  xhr.setRequestHeader('Accept', 'text/html, application/xhtml+xml, application/xml');
  xhr.setRequestHeader('X-XHR-Referer', document.location.href);
  xhr.send(req.data);

  // Events
  xhr.onload = function() {
  // @TODO: manage scroll position in here and stuff like that
  };

  xhr.onloadend = req.success;
  xhr.onerror = req.error;
}