$(document).ready(function () {

  $('.toggle-menu').jPushMenu({ closeOnClickLink: false });
  $('.dropdown-toggle').dropdown();

  $('pre code').each(function(i, block) {
    hljs.highlightBlock(block);
  });

});