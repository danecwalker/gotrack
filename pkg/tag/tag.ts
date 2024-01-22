interface Window {
  g?: (eventName: string, options?: object & { props?: object, callback?: () => void }) => void;
}

(() => {
  const document = window.document;
  const location = window.location;
  const screenSize = window.screen.width + 'x' + window.screen.height;
  const currentScript = document.currentScript as HTMLScriptElement;
  const api_url = currentScript.getAttribute('data-api') || getDefaultApiEndpoint(currentScript);
  
  function getDefaultApiEndpoint(script: HTMLScriptElement): string {
    return new URL(script.src).origin + '/e';
  }
  function ignoreEvent(reason: string, options: object & { callback?: () => void } = {}) {
     if (reason) console.info("Ignoring event: " + reason);
     options && options.callback && options.callback();
  }

  function sendEvent(eventName: string, options: object & { props?: object, callback?: () => void } = {}) {
    var payload: Partial<{
      n: string; // event name
      u: string; // url
      r?: string; // referrer
      p?: object; // props
      s?: string; // screen size
    }> = {};
    payload.n = eventName;
    payload.u = location.href;
    payload.r = document.referrer || undefined;
    if (options && options.props) {
      payload.p = options.props;
    }
    payload.s = screenSize;

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

  var lastpage: string;
  function pageView() {
    lastpage = location.pathname
    sendEvent('pageview')
  }

  var windowHistoryPushState = window.history.pushState;
  window.history.pushState = function(data: any, title: string, url?: string | URL | null | undefined) {
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
})();