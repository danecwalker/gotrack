{{define "custom"}}
var PARENT_LIMIT = 3;

function isForm(element) {
  return element && element.tagName && element.tagName.toLowerCase() === 'form'
}

function isOutboundLink(link) {
  return link && link.href && link.host && link.host !== location.host
}

function isLink(element) {
  return element && element.tagName && element.tagName.toLowerCase() === 'a'
}

function getTaggedEventAttr(element) {
  var eventAttr = {
    name: null,
    props: {}
  };
  {{- if .IncludeRevenue -}}
  eventAttr.$ = {};
  {{- end -}}

  var attrs = element.getAttributeNames && element.getAttributeNames() || [];
  for (var i = 0; i < attrs.length; i++) {
    if (attrs[i].indexOf('ga-event-') === 0) {
      var prop = attrs[i].substr(9);
      var value = element.getAttribute(attrs[i]);

      if (prop === 'name') { eventAttr.name = value; continue }
      eventAttr.props[prop] = value;
    }
    {{- if .IncludeRevenue -}}
    if (attrs[i].indexOf('ga-revenue-') === 0) {
      var prop = attrs[i].substr(11);
      var value = element.getAttribute(attrs[i]);
      eventAttr.$[prop] = value;
    }
    {{- end -}}
  }
  return eventAttr;
}

function followLink(event, link) {
  if (event.defaultPrevented) { return false }

  var targetsCurrentWindow = !link.target || link.target === '_self' || link.target === '_top' || link.target === '_parent';
  var regularClick = !(event.ctrlKey || event.shiftKey || event.metaKey) && event.type === 'click';
  return targetsCurrentWindow && regularClick
}

function sendLinkClickEvent(event, link, eventAttr) {
  var shouldFollow = false;
  function follow() {
    if (!shouldFollow) {
      shouldFollow = true;
      window.location = link.href;
    }
  }
  if (followLink(event, link)) {
    var attrs = { props: eventAttr.props, callback: follow };
    g(eventAttr.name, attrs);
    setTimeout(follow, 5000)
    event.preventDefault();
  } else {
    var attrs = { props: eventAttr.props }
    g(eventAttr.name, attrs)
  }
}

function handleTaggedClickEvent(event) {
  if (event.type === 'auxclick' && event.button !== 1) { return }

  var target = event.target;
  
  var clickedLink, taggedElement;
  for (var i = 0; i < PARENT_LIMIT; i++) {
    if (!target) { break }

    if (isForm(target)) { return }
    if (isLink(target)) { clickedLink = target }
    if (isTagged(target)) { taggedElement = target }
    target = target.parentNode;
  }

  if (taggedElement) {
    var eventAttr = getTaggedEventAttr(taggedElement);

    if (clickedLink) {
      eventAttr.props.url = clickedLink.href;
      sendLinkClickEvent(event, clickedLink, eventAttr);
    } else {
      var attr = {}
      attr.props = eventAttr.props;
      {{- if .IncludeRevenue -}}
      attr.$ = eventAttr.$;
      {{- end -}}
      g(eventAttr.name, attr)
    }
  } else {
    if (clickedLink && isOutboundLink(clickedLink)) {
      sendLinkClickEvent(event, clickedLink, { name: 'outboundlink', props: { url: clickedLink.href } })
    }
  }
}

function isTagged(element) {
  if (element && element.hasAttribute && element.hasAttribute('ga-event-name')) { return true }
  return false
}

document.addEventListener('click', handleTaggedClickEvent);
document.addEventListener('auxclick', handleTaggedClickEvent);
{{end}}