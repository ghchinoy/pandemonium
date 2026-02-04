var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
import { LitElement, html, css } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import './components/pan-agent-card';
import './components/pan-simulation-log';
import './components/pan-trace-inspector';
import './components/pan-world-state';
let PanObserverApp = class PanObserverApp extends LitElement {
    static { this.styles = css `
    :host { font-family: 'Inter', sans-serif; display: flex; flex-direction: column; height: 100vh; background: #0f0f0f; color: #eee; box-sizing: border-box; }
    header { padding: 15px 25px; border-bottom: 1px solid #222; display: flex; align-items: center; justify-content: space-between; flex-shrink: 0; }
    h1 { margin: 0; font-size: 1.2em; color: #7d56f4; letter-spacing: 1px; }
    .status-area { display: flex; align-items: center; gap: 20px; font-size: 0.8em; }
    .content { display: grid; grid-template-columns: 1fr 350px; gap: 20px; padding: 20px; flex: 1; min-height: 0; transition: grid-template-columns 0.3s ease-in-out; }
    .content.with-inspector { grid-template-columns: 1fr 350px 450px; }
    .sidebar { display: flex; flex-direction: column; gap: 10px; overflow-y: hidden; }
    .agent-list { overflow-y: auto; flex: 1; display: flex; flex-direction: column; gap: 10px; }
    .sidebar-header { font-size: 0.75em; font-weight: bold; color: #555; text-transform: uppercase; margin-bottom: 5px; }
    
    .toggle { display: flex; align-items: center; gap: 8px; cursor: pointer; color: #888; }
    .toggle input { display: none; }
    .toggle .slider { width: 30px; height: 16px; background: #333; border-radius: 10px; position: relative; }
    .toggle .slider::after { content: ''; position: absolute; width: 12px; height: 12px; background: #fff; border-radius: 50%; top: 2px; left: 2px; transition: 0.2s; }
    .toggle input:checked + .slider { background: #7d56f4; }
    .toggle input:checked + .slider::after { left: 16px; }
  `; }
    constructor() {
        super();
        this.events = [];
        this.agents = {};
        this.selectedTrace = null;
        this.currentStep = 0;
        this.premise = "";
        this.autoscroll = true;
        this.completed = false;
        this.connect();
    }
    connect() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const ws = new WebSocket(protocol + '//' + window.location.host + '/ws');
        ws.onmessage = (msg) => {
            const event = JSON.parse(msg.data);
            // Update simulation global state
            if (event.type === 'sim_start') {
                this.premise = event.data || event.message;
            }
            if (event.type === 'step_start') {
                this.currentStep = typeof event.data === 'number' ? event.data : this.currentStep;
            }
            if (event.type === 'sim_complete') {
                this.completed = true;
            }
            this.events = [...this.events, event];
            if (event.actor) {
                // Correctly handle registration metadata
                const eventIsRemote = event.data && typeof event.data === 'object' && event.data.isRemote === true;
                let existing = this.agents[event.actor];
                let currentStatus = existing?.status || 'Ready';
                let remoteStatus = existing?.isRemote || eventIsRemote;
                if (event.type === 'thought') {
                    currentStatus = 'Thinking...';
                }
                else if (event.type === 'observation') {
                    currentStatus = 'Observing';
                }
                else if (event.type === 'action') {
                    currentStatus = 'Acting';
                }
                this.agents = {
                    ...this.agents,
                    [event.actor]: {
                        status: currentStatus,
                        isRemote: !!remoteStatus
                    }
                };
            }
            if (event.type === 'resolution') {
                const newAgents = { ...this.agents };
                Object.keys(newAgents).forEach(name => {
                    newAgents[name] = { ...newAgents[name], status: 'Ready' };
                });
                this.agents = newAgents;
            }
        };
    }
    render() {
        return html `
      <header>
        <h1>PANDEMONIUM <span style="color: #444; font-weight: normal;">| OBSERVER</span></h1>
        <div class="status-area">
          <label class="toggle">
            <input type="checkbox" ?checked="${this.autoscroll}" @change="${(e) => this.autoscroll = e.target.checked}">
            <div class="slider"></div>
            Autoscroll
          </label>
          <div style="color: #04b575;">● SYSTEM LIVE</div>
        </div>
      </header>
      <main class="content ${this.selectedTrace ? 'with-inspector' : ''}" 
            @select-trace="${(e) => this.selectedTrace = e.detail}"
            @close-trace="${() => this.selectedTrace = null}">
        
        <pan-simulation-log .events="${this.events}" .autoscroll="${this.autoscroll}"></pan-simulation-log>
        
        <aside class="sidebar">
          <pan-world-state 
            .step="${this.currentStep}" 
            .premise="${this.premise}" 
            .completed="${this.completed}">
          </pan-world-state>
          <div class="sidebar-header">Agent Registry</div>
          <div class="agent-list">
            ${Object.entries(this.agents).map(([name, state]) => html `
              <pan-agent-card .name="${name}" .status="${state.status}" .isRemote="${state.isRemote}"></pan-agent-card>
            `)}
          </div>
        </aside>

        ${this.selectedTrace ? html `
          <pan-trace-inspector .trace="${this.selectedTrace}"></pan-trace-inspector>
        ` : ''}
      </main>
    `;
    }
};
__decorate([
    state(),
    __metadata("design:type", Array)
], PanObserverApp.prototype, "events", void 0);
__decorate([
    state(),
    __metadata("design:type", Object)
], PanObserverApp.prototype, "agents", void 0);
__decorate([
    state(),
    __metadata("design:type", Object)
], PanObserverApp.prototype, "selectedTrace", void 0);
__decorate([
    state(),
    __metadata("design:type", Object)
], PanObserverApp.prototype, "currentStep", void 0);
__decorate([
    state(),
    __metadata("design:type", Object)
], PanObserverApp.prototype, "premise", void 0);
__decorate([
    state(),
    __metadata("design:type", Object)
], PanObserverApp.prototype, "autoscroll", void 0);
__decorate([
    state(),
    __metadata("design:type", Object)
], PanObserverApp.prototype, "completed", void 0);
PanObserverApp = __decorate([
    customElement('pan-observer-app'),
    __metadata("design:paramtypes", [])
], PanObserverApp);
export { PanObserverApp };
//# sourceMappingURL=pan-observer-app.js.map