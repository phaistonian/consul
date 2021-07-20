import Application from '../app';
import config from '../config/environment';
import { setApplication } from '@ember/test-helpers';
import { registerWaiter } from '@ember/test';
import { getOwner } from '@ember/application';
import './helpers/flash-message';
import start from 'ember-exam/test-support/start';

import ClientConnections from 'consul-ui/services/client/connections';
import LocalStorageClipboard from 'consul-ui/services/clipboard/local-storage';

let activeRequests = 0;
registerWaiter(function() {
  return activeRequests === 0;
});
ClientConnections.reopen({
  addVisibilityChange: function() {
    // for the moment don't listen for tab hiding during testing
    // TODO: make this controllable from testing so we can fake a tab hide
  },
  purge: function() {
    const res = this._super(...arguments);
    activeRequests = 0;
    return res;
  },
  acquire: function() {
    activeRequests++;
    return this._super(...arguments);
  },
  release: function() {
    const res = this._super(...arguments);
    activeRequests--;
    return res;
  },
});
const application = Application.create(config.APP);
application.register('service:clipboard', LocalStorageClipboard);
setApplication(application);

start();
