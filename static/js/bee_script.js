/**
 * Script.js
 */
// jQuery(function () {
$(document).ready(function () {
  var $bee_view = $('#bee_view').val(); // = nom du cookie

  $(".draggable").draggable({
    revert: true,
    opacity: 0.5,
    helper: 'clone'
  }); // glisser
  $(".droppable").droppable({
    classes: {
      "ui-droppable-active": "teal",
      "ui-droppable-hover": "active"
    },
    drop: function (event, ui) {
      var $dest = $(this).data('bdir');
      var $dsrc = ui.draggable.data('dsrc');
      var $fsrc = ui.draggable.data('fsrc');
      // console.log('drag', $dsrc, $fsrc, 'drop', $dest);
      var $form = $('#bee-drag-drop').find('form');
      $form.attr('action', '/dragdrop/' + $dest);
      $form.find("input[name='dsrc']").val($dsrc);
      $form.find("input[name='fsrc']").val($fsrc);
      $form.submit();
    }
  });

  $('.bee-nag').on('tap', function (event) {
    $('#bee-nag')
      .nag({
        persist: true
      })
  });

  $('.bee-meta').on('tap', function (event) {
    var $card = $(this).closest('.card');
    var $anchorid = $card.attr('id');
    Cookies.set($bee_view, $anchorid)
    // $('.bee-selected').removeClass('bee-selected');
    window.open($(this).data("url"), "_self");
    event.preventDefault();
  });

  $('.bee-onchange').on('change', function (event) {
    $(".bee-submit-meta").removeClass('disabled');
    event.preventDefault();
  });

  // SELECTION MULTIPLE
  $('#bee-selector').on('tap', function (event) {
    $(this).html('<i class="check icon"></i>')
    $('.bee-selected').removeClass('bee-selected');
    $('.bee-press-visible').hide();
    event.preventDefault();
  });

  // Sélection d'un dossier ou fichier
  $('.bee-press').on('press', function (event) {
    if ($(this).hasClass('bee-selected')) {
      // désélection
      $(this).removeClass('bee-selected');
      if ($('.bee-selected').length == 0)
        $('.bee-press-visible').hide();
      else {
        $('#bee-selector').html($('.bee-selected').length);
      }
    } else {
      // sélection
      // $(this).parent().find('.bee-selected').removeClass('bee-selected');
      $(this).addClass("bee-selected");
      $('.bee-press-visible').show();
      $('#bee-selector').html($('.bee-selected').length);
    }
    event.preventDefault();
  });

  $('.bee-submit-meta').on('tap', function (event) {
    var $form = $('#form_meta_id');
    $form.submit();
    event.preventDefault();
  });

  // ACTION NEW
  $('.bee-modal-new').on('tap', function (event) {
    var $form = $('#bee-modal-new').find('form');
    $form.attr('action', $(this).data('action'));
    $('#bee-modal-new').find('.header').html($(this).attr('title'));
    $('#bee-modal-new').find("input[name='new_name']").val('');
    $('#bee-modal-new')
      .modal({
        closable: false,
        onDeny: function () {
          return true;
        },
        onApprove: function () {
          $form.submit();
        }
      }).modal('show');
    event.preventDefault();
  });
  // ACTION RENAME
  $('.bee-modal-rename').on('tap', function (event) {
    var $form = $('#bee-modal-rename').find('form');
    // valorisation de paths et bases
    $form.attr('action', $(this).data('action'));
    $('#bee-modal-rename').find('.bee-modal-title').html($(this).attr('title'));
    $('#bee-modal-rename').find("input[name='new_name']").attr('placeholder', $(this).data('default'));
    $('#bee-modal-rename').find("input[name='new_name']").val($(this).data('default'));
    $('#bee-modal-rename')
      .modal({
        closable: false,
        onDeny: function () {
          return true;
        },
        onApprove: function () {
          $form.submit();
        }
      }).modal('show');
    event.preventDefault();
  });

  // ACTION DOWNLOAD
  $('.bee-select-download').on('tap', function (event) {
    // Recherche du fichier sélectionné qui sera unique
    $selected = getSelectedPathHtml();
    var link = document.createElement('a');
    link.href = $selected.paths;
    link.download = $selected.baseUnique;
    link.click();
    // window.open($selected.paths, '_blank');
    event.preventDefault();
  });

  // ACTION CONFIRMATION
  $('.bee-modal-confirm').on('tap', function (event) {
    var $form = $('#bee-modal-confirm').find('form');
    $('.bee-modal-title').html($(this).attr('title'));
    $form.attr('action', $(this).data('action'));
    if ($(this).data('message')) {
      $('#bee-modal-confirm').find('.message>.header').html($(this).data('message'));
    }
    $('#bee-modal-confirm')
      .modal({
        closable: false,
        onDeny: function () {
          return true;
        },
        onApprove: function () {
          $('form', document).submit();
        }
      }).modal('show');
    event.preventDefault();
  });
  // ACTION RESTAURATION
  $('.bee-modal-restore').on('tap', function (event) {
    var $form = $('#bee-modal-restore').find('form');
    $('#bee-modal-restore')
      .modal({
        closable: false,
        onDeny: function () {
          return true;
        },
        onApprove: function () {
          $form.submit();
        }
      }).modal('show');
    event.preventDefault();
  });
  // ACTION UPLOAD
  // affichage de la fenêtre modal
  $('.bee-modal-upload').on('tap', function (event) {
    var $form = $('#bee-modal-upload').find('form');
    $('#bee-modal-upload')
      .modal({
        closable: false,
        onDeny: function () {
          return true;
        },
        onApprove: function () {
          $form.submit();
        }
      }).modal('show');
    event.preventDefault();
  });
  // maj des fichier à télécharger dans la zone #bee-upload-file
  $('#bee-upload-file').on('change', function () {
    var $files = $(this).get(0).files;
    var $html = "";
    for (var i = 0; i < $files.length; i++) {
      var $filename = $files[i].name.replace(/.*(\/|\\)/, '');
      $html += '<div class="ui teal label">' + $filename + '</div>'
    }
    $('#bee-files-selected').html($html);
  });

  // ACTION DUPLIQUER
  $('.bee-modal-duplicate').on('click', function (event) {
    var $modal = $('#bee-modal-duplicate')
    // titre
    $modal.find('.bee-modal-title').html($(this).attr('title'));
    var $form = $modal.find('form');
    // valorisation de paths et bases
    $selected = getSelectedPathHtml();
    // Le champ input des fichiers sources
    $form.find('input[name="paths"]').val($selected.paths)
    $form.find('.bee-input-paths').html($selected.bases)
    // l'action à déclencher sur le serveur
    $form.attr('action', $(this).data('action'));
    $('#bee-modal-duplicate')
      .modal({
        closable: false,
        onDeny: function () {
          return true;
        },
        onApprove: function () {
          $form.submit();
        }
      }).modal('show');
    event.preventDefault();
  });

  // ACTION COPIER ou DEPLACER
  $('.bee-modal-move').on('click', function (event) {
    var $modal = $('#bee-modal-move')
    // titre
    $modal.find('.bee-modal-title').html($(this).attr('title'));
    var $form = $modal.find('form');
    // valorisation de paths et bases
    $selected = getSelectedPathHtml();
    // Le champ input des fichiers sources
    $form.find('input[name="paths"]').val($selected.paths)
    $form.find('.bee-input-paths').html($selected.bases)
    // Le champ input du répertoire destination par défaut
    var $folder = $('#bee-ctx').data('folder')
    $form.find('input[name="dest"]').val($folder)
    $form.find('.bee-input-dest').html($folder)
    // l'action à déclencher sur le serveur
    $form.attr('action', $(this).data('action'));
    $('#bee-modal-move')
      .modal({
        closable: false,
        onDeny: function () {
          return true;
        },
        onApprove: function () {
          $form.submit();
        }
      }).modal('show');
    event.preventDefault();
  });

  // ACTION SUPPRIMER
  $('.bee-modal-delete').on('tap', function (event) {
    var $modal = $('#bee-modal-confirm')
    // titre
    $modal.find('.bee-modal-title').html($(this).attr('title'));
    var $form = $modal.find('form');
    // valorisation de paths et bases
    $selected = getSelectedPathHtml();
    // Le champ input des fichiers sources
    $form.find('input[name="paths"]').val($selected.paths)
    // l'action à déclencher sur le serveur
    $form.attr('action', $(this).data('action'));
    // le message dans la modal
    $form.find('.message>.header').html($selected.bases);
    $('#bee-modal-confirm')
      .modal({
        closable: false,
        onDeny: function () {
          return true;
        },
        onApprove: function () {
          $('form', document).submit();
        }
      }).modal('show');
    event.preventDefault();
  });

  // CLIC IMAGE EDITOR POPUP
  $('.bee-popup-image-editor').on('tap', function (event) {
    var $form = $('#form_meta_id');
    var $image = $form.find('img');
    var $url = $(this).data('src');
    var $input = $form.find("input[name='image']");
    const config = {
      language: 'fr',
      tools: ['adjust', 'effects', 'filters', 'rotate', 'crop', 'resize', 'text'],
      colorScheme: 'dark',
      translations: {
        fr: {
          'toolbar.download': 'Valider'
        },
      }
    };
    var mime = $url.endsWith('.png') ? 'image/png' : 'image/jpeg';
    // https://github.com/scaleflex/filerobot-image-editor
    const ImageEditor = new FilerobotImageEditor(config, {
      onBeforeComplete: (props) => {
        // console.log("onBeforeComplete", props);
        // console.log("canvas-id", props.canvas.id);
        var canvas = document.getElementById(props.canvas.id);
        var dataurl = canvas.toDataURL(mime, 1);
        // update image du browser
        $image.attr('src', dataurl);
        // remplissage du imput pour le submit
        $input.val(dataurl);
        $(".bee-submit-meta").removeClass('disabled');
        return false;
      },
      onComplete: (props) => {
        // console.log("onComplete", props);
        return true;
      }
    });
    ImageEditor.open($url);
    event.preventDefault();
  });

  // IHM SEMANTIC
  // $('.menu .item').tab();
  $('.ui.checkbox').checkbox();
  $('.ui.radio.checkbox').checkbox();
  $('.ui.dropdown.item').dropdown();
  $('select.dropdown').dropdown();
  $('.message .close')
    .on('click', function () {
      $(this)
        .closest('.message')
        .transition('fade');
    });

  // Toaster
  $('#toaster')
    .toast({
      class: $('#toaster').data('color'),
      position: $('#toaster').data('position'),
      message: $('#toaster').val()
    });

  /**
   * Ouverture d'une fenêtre en popup
   */
  $(document).on('tap', '.bee-window-open', function (event) {
    // Préparation window.open
    var height = $(this).data("height") ? $(this).data("height") : 'max';
    var width = $(this).data("width") ? $(this).data("width") : 'large';
    var posx = $(this).data("posx") ? $(this).data("posx") : 'left';
    var posy = $(this).data("posy") ? $(this).data("posy") : '3';
    var target = $(this).attr("target") ? $(this).attr("target") : 'Bee-win';
    if (window.opener == null) {
      window.open($(this).data('url'), target, computeWindow(posx, posy, width, height, false));
    } else {
      window.opener.open($(this).data('url'), target, computeWindow(posx, posy, width, height, false));
    }
    event.preventDefault();
  });

  /**
   * Fermeture de la fenêtre popup
   */
  $(document).on('tap', '.bee-confirm-close', function (event) {
    if ($('#button_validate').length > 0 &&
      $('#button_validate').hasClass('disabled') == false) {
      var $url = $(this).data('url')
      $('#bee-modal-confirm')
        .modal({
          closable: false,
          onDeny: function () {
            return true;
          },
          onApprove: function () {
            if ($url) {
              window.open($url, '_self');
            } else {
              window.close();
            }
          }
        }).modal('show');
    } else {
      var $url = $(this).data('url')
      if ($url) {
        window.open($url, '_self');
      } else {
        window.close();
      }
    }
    event.preventDefault();
  });

  /**
   * retourne le HTML des chemins concaténés des fichiers et répertoires sélectionnés
   */
  function getSelectedPathHtml() {
    // valorisation de bee-path
    var $paths = ""; var $bases = ""; var $baseUnique = ""
    $('.bee-selected').each(function () {
      if ($paths.length > 0) {
        $paths += ",";
      }
      $paths += $(this).data('path');
      $bases += '<span class="ui teal label">' + $(this).data('base') + '</span>';
      $baseUnique = $(this).data('base');
    });
    return {
      paths: $paths, bases: $bases, baseUnique: $baseUnique
    }
  }

  // APPEL DRAWIO
  $('.bee-drawio').on('tap', function (event) {
    var url = 'https://embed.diagrams.net/?embed=1&ui=atlas&spin=1&modified=unsavedChanges&proto=json';
    var source = $('#bee-drawio')[0];
    // var title = source.getAttribute('title')
    // url += '&title=' + title;
    if (source.drawIoWindow == null || source.drawIoWindow.closed) {
      // Implements protocol for loading and exporting with embedded XML
      var receive = function (evt) {
        if (evt.data.length > 0 && evt.source == source.drawIoWindow) {
          var msg = JSON.parse(evt.data);

          // Received if the editor is ready
          if (msg.event == 'init') {
            // Sends the data URI with embedded XML to editor
            source.drawIoWindow.postMessage(JSON.stringify(
              { action: 'load', xmlpng: source.getAttribute('src') }), '*');
          }
          // Received if the user clicks save
          else if (msg.event == 'save') {
            // Sends a request to export the diagram as XML with embedded PNG
            source.drawIoWindow.postMessage(JSON.stringify(
              { action: 'export', format: 'xmlpng', spinKey: 'saving' }), '*');
          }
          // Received if the export request was processed
          else if (msg.event == 'export') {
            // Updates the data URI of the image
            source.setAttribute('src', msg.data);
            $('input[name="image"]').val(msg.data);
            $(".bee-submit").removeClass('disabled');
          }

          // Received if the user clicks exit or after export
          if (msg.event == 'exit' || msg.event == 'export') {
            // Closes the editor
            window.removeEventListener('message', receive);
            source.drawIoWindow.close();
            source.drawIoWindow = null;
          }
        }
      };
      // Opens the editor
      window.addEventListener('message', receive);
      var $height = 'max';
      var $width = 'max';
      var $posx = '5';
      var $posy = '5';
      var $target = '_blank';
      source.drawIoWindow = window.open(url, $target, computeWindow($posx, $posy, $width, $height, false));
      // source.drawIoWindow = window.open(url);
    }
    else {
      // Shows existing editor window
      source.drawIoWindow.focus();
    }
  });
  // paramètres de Lighbox
  lightbox.option({
    'fitImagesInViewport': true,
    'alwaysShowNavOnTouchDevices': true,
    'wrapAround': true
  })

  // chargement de toutes les images avant de lancer Masonry
  Promise.all(Array.from(document.images)
    .filter(img => !img.complete)
    .map(img => new Promise(resolve => { img.onload = img.onerror = resolve; }))).then(() => {
      var $grid = $('.grid').masonry({
        itemSelector: '.grid-item'
        // horizontalOrder: true
      });
      $grid.masonry();
      // positionnement sur la dernière carte sélectionnée
      // sélection clic sur metadata
      if ($bee_view && $bee_view.length > 0) {
        if (Cookies.get($bee_view)) {
          var $anchor = $('#' + Cookies.get($bee_view));
          if ($anchor.length) {
            $('html, body').animate({
              scrollTop: $anchor.offset().top - 100
            }, 1000);
            // encadremant de la diapo
            $anchor.addClass("bee-card-anchor");
            // $anchor.css("border", "3px");
            // positionnement sur le menu sélectionné
            $('#menu').animate({
              scrollTop: $('.active').offset().top - 100
            }, 1000);
          }
        }
      };
    });
});

/**
 * Calcul du positionnement et de la taille de la fenêtre sur l'écran
 * @param {string} posx left center right ou px
 * @param {int} posy px
 * @param {string} pwidth max large xlarge ou px
 * @param {string} pheight max ou px
 * @param {srtien} full_screen yes no 0 1
 */
function computeWindow(posx, posy, pwidth, pheight, full_screen) {
  if (full_screen) {
    pheight = screen.availHeight - 70;
    pwidth = screen.availWidth - 6;
  }
  var height = pheight != null ? (/^max$/gi.test(pheight) ? screen.availHeight - 120 : pheight) : 830;
  var width = 900;
  if (pwidth != null) {
    width = pwidth;
    if (/^max$/gi.test(pwidth)) width = screen.availWidth - 6;
    if (/^large$|^l$/gi.test(pwidth)) width = 1024;
    if (/^xlarge$|^xl$/gi.test(pwidth)) width = 1248;
  } // end largeur
  var left = 3;
  if (posx != null) {
    left = posx;
    if (/^left$/gi.test(posx)) left = 3;
    if (/^right$/gi.test(posx)) left = screen.availWidth - width - 18;
    if (/^center$/gi.test(posx)) left = (screen.availWidth - width) / 2;
  } // end posx
  var top = 6
  if (posy != null) {
    height = screen.availHeight - posy - 10;
    top = posy;
  }

  return 'left=' + left + ',top=' + top + ',height=' + height + ',width=' + width + ',scrolling=yes,scrollbars=yes,resizeable=yes';
}

/**
 * Blocage du carriage return dans un champ input par exemple
 * @param {object event} event
 */
function blockCR(event) {
  if (event.keyCode == 13) {
    event.preventDefault();
  }
}
