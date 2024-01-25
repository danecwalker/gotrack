{{define "tag"}}
(function() {
  const document = window.document;
  const location = window.location;
  const currentScript = document.currentScript;
  const api_url = currentScript.getAttribute('data-api') || getDefaultApiEndpoint(currentScript);
  
  function getDefaultApiEndpoint(script) {
    return new URL(script.src).origin + '/e';
  }

  function ignoreEvent(reason, options) {
     if (reason) console.warn("Ignoring event: " + reason);
     options && options.callback && options.callback();
  }

  function sendEvent(eventName, options) {
    {{- if not .IsDebug -}}
    if (/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(location.hostname) || location.protocol === 'file:') {
      ignoreEvent('debug');
      return;
    }
    {{- end -}}
    var payload = {};
    payload.n = eventName;
    payload.u = location.href;
    payload.r = document.referrer || undefined;
    if (options && options.props) {
      payload.p = options.props;
    }
    payload.s = window.innerWidth + 'x' + window.innerHeight;
    {{- if .IncludeRevenue -}}
    if (options && options.$) {
      payload.$ = options.$;
    }
    {{- end -}}

    const request = new XMLHttpRequest();
    request.open('POST', api_url, true);
    request.setRequestHeader('Content-Type', 'application/json');
    request.send(JSON.stringify(payload));
    request.onreadystatechange = function() {
      if (request.readyState === 4) {
        if (request.status !== 202) {
          console.error("Error sending event: " + request.status);
        } else {
          options && options.callback && options.callback();
        }
      }
    };
  }

  window.g = window.g || sendEvent;

  var lastpage;
  function pageView() {
    lastpage = location.pathname
    sendEvent('pageview')
  }

  var windowHistoryPushState = window.history.pushState;
  window.history.pushState = function(data, title, url) {
    windowHistoryPushState.apply(this, [data, title, url]);
    pageView();
  }
  window.addEventListener('popstate', pageView);

  if (document.visibilityState !== 'visible') {
    document.addEventListener('visibilitychange', function() {
      if (!lastpage && document.visibilityState === 'visible') {
        pageView()
      }
    })
  } else {
    pageView()
  }

  {{- if .IncludeAll -}}
  {{template "custom" .}}
  {{- end -}}
})();
{{end}}