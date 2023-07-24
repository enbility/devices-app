<script setup>
import imageUrl from '../assets/logo_white.svg'
</script>
<template>
  <nav
    class="navbar is-fresh is-transparent no-shadow"
    role="navigation"
    aria-label="main navigation"
  >
    <div class="container">
      <div class="navbar-start">
        <div class="navbar-item">
          <h1 class="title has-text-white">Devices App</h1>
        </div>
      </div>
      <div class="navbar-end">
        <div class="navbar-item">
          <h5 class="subtitle is-5 has-text-white" v-show="isOnline">SKI: {{ serviceSki }}</h5>
        </div>
      </div>
    </div>
  </nav>

  <div class="container mt-50">
    <div class="content has-text-center" v-show="!isOnline">
      <div class="notification is-danger is-light">
        Server connection is lost. Please check that Devices-App is still running.
      </div>
    </div>

    <div class="content has-text-right" v-if="isOnline">
      <label class="checkbox">
        <input type="checkbox" v-model="allowRemote">
        Allow Remote Connections
      </label>
    </div>
    
    <div class="tabs" v-show="isOnline">
      <ul>
        <li :class="{ 'is-active': currentTab == 1 }">
          <a @click="showServices()">1. Select EEBUS service</a>
        </li>
        <li :class="{ 'is-active': currentTab == 2 }">
          <a :class="{ 'is-disabled': visibleServiceSki == null || visibleServiceState == ConnectionStateEnum.COMPLETED }">
            2. Initiate communication
          </a>
        </li>
        <li :class="{ 'is-active': currentTab == 3 }">
          <a :class="{ 'is-disabled': visibleServiceSki == null || visibleServiceState != ConnectionStateEnum.COMPLETED }">
            3. View capabilities
          </a>
        </li>
      </ul>
    </div>

    <div class="columns" id="tab1" v-if="currentTab == 1 && isOnline">
      <div class="column">
        <table class="table is-fullwidth">
          <thead>
            <tr>
              <th>Brand & Model</th>
              <th>SKI</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in externalServices">
              <td class="has-text-left">
                <a @click="selectService(item.ski)">
                  {{ item.brand }} {{ item.model }}
                </a>
              </td>
              <td class="has-text-left">
                <small>{{ item.ski }}</small>
              </td>
              <td>
                <div class="buttons">
                  <template v-if="item.state == ConnectionStateEnum.INITIATED || item.state == ConnectionStateEnum.INPROGRESS">
                    <button 
                      class="button is-warning is-small"
                      @click="abortConnection(item.ski)"
                    >
                      <template v-if="item.trusted">
                        Cancel connection
                      </template>
                      <template v-if="!item.trusted">
                        Cancel pairing
                      </template>
                    </button>
                  </template>
                  <template v-if="item.incomingRequest">
                    <button 
                      class="button is-info is-small"
                      @click="pairService(item.ski)"
                    >
                      Accept pairing request
                    </button>
                    <button 
                      class="button is-danger is-small"
                      @click="abortConnection(item.ski)"
                    >
                      Deny pairing request
                    </button>
                  </template>
                  <button 
                    class="button is-danger is-small"
                    @click="unpairService(item.ski)"
                    v-show="item.trusted || item.state == ConnectionStateEnum.COMPLETED"
                  >
                    Remove pairing
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="columns" id="breadcrumb" v-if="currentTab > 1 && isOnline">
      <nav class="breadcrumb" aria-label="breadcrumbs">
        <ul>
          <li class="is-active">
            <a aria-current="page">{{ visibleServiceTitle }}</a>
            <small>({{ visibleServiceSki }})</small>
          </li>
        </ul>
      </nav>
    </div>

    <div class="columns" id="tab2" v-if="currentTab == 2 && isOnline">
      <div class="column">
        <template v-if="visibleServiceIncoming">
          <div class="content has-text-centered">
            <p>
              The remote service  <strong>{{ visibleServiceTitle }}</strong>
              wants to communicate with this <strong>Desktop App</strong>.
              To allow this, it needs to be paired first.
            </p>
            <div class="buttons" style="display:block !important">
              <button 
                class="button is-info is-small"
                @click="pairService(visibleServiceSki)"
              >
                Accept pairing request
              </button>
              <button 
                class="button is-danger is-small"
                @click="abortConnection(visibleServiceSki)"
              >
                Deny pairing request
              </button>
            </div>
          </div>
        </template>
        <template v-if="!visibleServiceIncoming">
          <div
            class="content has-text-centered"
            v-if="visibleServiceState == ConnectionStateEnum.NONE">
            <p>
              To be able to communicate with
              <strong>{{ visibleServiceTitle }}</strong>
              , it needs to be paired first.
            </p>
            <button
              class="button is-primary"
              @click="pairService(visibleServiceSki)"
            >
              Start Pairing Process
            </button>
          </div>
          <div
            class="content has-text-centered"
            v-else-if="visibleServiceState == ConnectionStateEnum.ERROR"
          >
            <div class="notification is-danger is-light">
              <p>
                Connecting ended in an error:
              <strong>{{ visibleServiceError }}</strong>
              </p>
              <p>
                Do you want to try it again?
              </p>
            </div>
            <button
              class="button is-primary"
              @click="pairService(visibleServiceSki)"
            >
              Start Pairing Process
            </button>
          </div>
          <div
            class="content has-text-centered"
            v-else-if="visibleServiceState != ConnectionStateEnum.REMOTEDENIEDTRUST"
          >
            <progress class="progress is-large is-info" max="100"></progress>
            <p>
              Pairing is in progress ...
            </p>
            <button class="button is-warning" @click="abortConnection(visibleServiceSki)">
              Cancel Pairing Process
            </button>
          </div>
          <div
            class="content"
            v-else-if="visibleServiceState == ConnectionStateEnum.REMOTEDENIEDTRUST"
          >
            <div class="notification is-warning is-light">
              <p>
                <strong>{{ visibleServiceTitle }}</strong> denied
                communication with this <strong>Desktop App</strong>!
              </p>
              <p>
                Please open the web interface of
                <strong>{{ visibleServiceTitle }}</strong>
                and accept pairing it with this
                <strong>Desktop App</strong>
                after starting the pairing process again.
              </p>
            </div>
            <div class="has-text-centered">
              <button
                class="button is-primary"
                @click="pairService(visibleServiceSki)"
              >
                Start Pairing Process
              </button>
            </div>
          </div>
        </template>
      </div>
    </div>

    <div class="columns" id="tab3" v-if="currentTab == 3 && isOnline">
      <div class="column">
        <div class="content">
          <p>
            <button 
              class="button is-danger is-small"
              @click="unpairService(visibleServiceSki)"
            >
              Remove pairing
            </button>
          </p>
          <p>
            <strong>Discovery Data:</strong>
            <pre>{{ visibleServiceDiscovery }}</pre>
          </p>
          <p>
            <strong>Usecase Data:</strong>
            <pre>{{ visibleServiceUsecase }}</pre>
          </p>
        </div>
      </div>
    </div>
</div>
</template>

<script>
import { ConnectionStateEnum } from "./../consts.js";

export default {
  name: "Devices-App",
  data: function () {
    return {
      ConnectionStateEnum: ConnectionStateEnum,
      isOnline: false,
      ws: null,
      reconnectTimeout: null,
      services: [],
      discoveryData: [],
      visibleServiceSki: null,
      allowRemote: true,
    };
  },

  created: function () {
    var self = this;

    this.wsConnect();
  },

  watch: {
    allowRemote(newValue, oldValue) {
      if (newValue == oldValue) {
        return;
      }

      var msg = {
        name: "allowremote",
        enable: newValue,
      };
      this.ws.send(JSON.stringify(msg));
    }
  },

  computed: {

    currentTab() {
      if (this.visibleServiceSki === null) {
        return 1;
      }

      if (this.visibleServiceState == ConnectionStateEnum.COMPLETED) {
        return 3;
      }

      return 2;
    },

    serviceSki() {
      var service = this.services.filter(item => item.itself);
      if (service.length > 0) {
        return service[0].ski;
      }
      return "";
    },

    externalServices() {
      return this.services.filter(item => !item.itself);
    },

    visibleServiceIncoming() {
      var service = this.serviceForSki(this.visibleServiceSki);

      if (service === false) {
        return false;
      }

      return service.incomingRequest;

    },

    visibleServiceState() {
      var service = this.serviceForSki(this.visibleServiceSki);

      if (service === false) {
        return ConnectionStateEnum.NONE;
      }

      return service.state;
    },

    visibleServicePaired() {
      var service = this.serviceForSki(this.visibleServiceSki);

      if (service === false) {
        return false;
      }

      return service.paired;
    },

    visibleServiceTitle() {
      var service = this.serviceForSki(this.visibleServiceSki);

      if (service === false) {
        return "not found";
      }

      return service.brand + " " + service.model;
    },

    visibleServiceDiscovery() {
      var service = this.serviceForSki(this.visibleServiceSki);

      if (service === false) {
        return "not found";
      }

      var json = JSON.parse(service.discovery);
      return json;
    },

    visibleServiceUsecase() {
      var service = this.serviceForSki(this.visibleServiceSki);

      if (service === false) {
        return "not found";
      }

      var json = JSON.parse(service.usecase);
      return json;
    },

    visibleServiceError() {
      if (this.visibleServiceSki == "") {
        return "";
      }
      return this.pairingErrorForSki(this.visibleServiceSki);
    },
  },

  methods: {
    wsConnect: function () {
      if (this.ws) {
        return;
      }
      const location = window.location;
      const protocol = location.protocol == "https:" ? "wss" : "ws:";
      const uri =
        protocol +
        "//" +
        location.hostname +
        (location.port ? ":" + location.port : "") +
        location.pathname +
        "ws";
      this.ws = new WebSocket(uri);

      this.ws.onmessage = this.wsOnMessage;
      this.ws.onopen = this.wsOnOpen;
      this.ws.onerror = this.wsOnError;
      this.ws.onclose = this.wsOnClose;
    },
    wsDisconnect: function () {
      if (!this.ws) {
        return;
      }
      this.ws.onmessage = null;
      this.ws.onopen = null;
      this.ws.onerror = null;
      this.ws.onclose = null;
      this.ws.close();
      this.ws = null;
    },
    wsReconnect: function () {
      self = this;
      window.clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = window.setTimeout(function() {
        self.wsDisconnect();
        self.wsConnect();
      }, 2000);
    },

    wsOnMessage: function (event) {
      var message = JSON.parse(event.data);

      switch (message.name) {
        case "serviceslist":
          this.services = message.services;

          this.updateIncomingRequests();
          break;

        case "allowremote":
          this.allowRemote = message.enable;
          break;

        default:
          break;
      }
    },
    wsOnOpen: function () {
      this.isOnline = true;

      var msg = { name: "serviceslist" };
      this.ws.send(JSON.stringify(msg));
    },
    wsOnClose: function () {
      this.isOnline = false;
      this.wsReconnect();
    },
    wsOnError: function () {
      this.ws.close();
    },

    showServices: function () {
      this.visibleServiceSki = null;
    },

    selectService: function (ski) {
      this.visibleServiceSki = ski;
    },

    pairService: function (ski) {
      var msg = {
        name: "pair",
        ski: ski,
      };
      this.ws.send(JSON.stringify(msg));
      // if (this.visibleServiceSki == ski) {
      //   this.visibleServiceState = ConnectionStateEnum.INITIATED;
      // }

      // this.selectService(ski);
    },

    abortConnection: function (ski) {
      var msg = {
        name: "abort",
        ski: ski,
      };
      this.ws.send(JSON.stringify(msg));
      // if (this.visibleServiceSki == ski) {
      //   this.visibleServiceState = ConnectionStateEnum.NONE;
      // }
      // this.showServices()
    },

    unpairService: function (ski) {
      var msg = {
        name: "unpair",
        ski: ski,
      };
      this.ws.send(JSON.stringify(msg));
      // if (this.visibleServiceSki == ski) {
      //   this.visibleServiceState = ConnectionStateEnum.NONE;
      // }
      // this.showServices()
    },

    serviceForSki: function (ski) {
      if (ski === null || ski === undefined || this.services === null || this.services === undefined) {
        return false;
      }

      for (let index = 0; index < this.services.length; index++) {
        const element = this.services[index];

        if (element.ski == ski) {
          return element;
        }
      }

      return false;
    },

    pairingErrorForSki: function (ski) {
      var service = this.serviceForSki(ski);
      if (service == false) {
        return "";
      }

      if (service.error.length > 0) {
        return service.error;
      }

      return "";
    },

    updateIncomingRequests: function () {
      for (let index = 0; index < this.services.length; index++) {
        const element = this.services[index];
        var incoming = false
        if (element.state === ConnectionStateEnum.RECEIVED) {
          incoming = true
        }
        this.services[index].incomingRequest = incoming
      }
    },

  },
};
</script>
