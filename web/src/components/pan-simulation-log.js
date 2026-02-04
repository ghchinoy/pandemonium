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
import { customElement, property } from 'lit/decorators.js';
let PanSimulationLog = class PanSimulationLog extends LitElement {
    constructor() {
        super(...arguments);
        this.events = [];
        this.autoscroll = true;
    }
    static { this.styles = css `
    :host { display: block; background: #000; padding: 20px; overflow-y: auto; border-radius: 12px; border: 1px solid #333; height: 100%; box-sizing: border-box; }
    .event { margin-bottom: 16px; padding-left: 12px; border-left: 3px solid #444; position: relative; }
    .type { font-family: 'JetBrains Mono', monospace; font-size: 0.7em; color: #666; margin-left: 8px; vertical-align: middle; }
    .actor { font-weight: bold; font-size: 0.85em; text-transform: uppercase; letter-spacing: 0.5px; }
    .message { margin-top: 4px; line-height: 1.5; color: #ccc; font-size: 0.95em; white-space: pre-wrap; }
    .inspect-btn { position: absolute; right: 0; top: 0; font-size: 0.7em; background: #222; border: 1px solid #333; color: #7d56f4; border-radius: 4px; padding: 2px 6px; cursor: pointer; opacity: 0; transition: opacity 0.2s; }
    .event:hover .inspect-btn { opacity: 1; }
    
    .observation { border-left-color: #04b575; }
    .observation .actor { color: #04b575; }
    .action { border-left-color: #ffd700; background: rgba(255, 215, 0, 0.03); padding: 8px 12px; border-radius: 0 4px 4px 0; cursor: pointer; }
    .action:hover { background: rgba(255, 215, 0, 0.06); }
    .action .actor { color: #ffd700; }
    .resolution { border-left-color: #ee6ff8; font-style: italic; }
    .resolution .actor { color: #ee6ff8; }
    .thought { border-left-color: #555; opacity: 0.6; font-size: 0.9em; }
    .thought .actor { color: #555; }
  `; }
    updated(changedProperties) {
        if (this.autoscroll && changedProperties.has('events')) {
            this.scrollTop = this.scrollHeight;
        }
    }
    render() {
        const displayEvents = this.events.filter(e => e.type !== 'step_start' && e.type !== 'sim_start');
        return html `
      ${displayEvents.map(e => html `
        <div class="event ${e.type}" @click="${() => this.handleEventClick(e)}">
          <span class="actor">${e.actor || 'World'}</span>
          <span class="type">(${e.type})</span>
          ${e.data ? html `<button class="inspect-btn">Inspect</button>` : ''}
          <div class="message">${e.message}</div>
        </div>
      `)}
    `;
    }
    handleEventClick(e) {
        if (e.data) {
            this.dispatchEvent(new CustomEvent('select-trace', {
                detail: e,
                bubbles: true,
                composed: true
            }));
        }
    }
};
__decorate([
    property({ type: Array }),
    __metadata("design:type", Array)
], PanSimulationLog.prototype, "events", void 0);
__decorate([
    property({ type: Boolean }),
    __metadata("design:type", Object)
], PanSimulationLog.prototype, "autoscroll", void 0);
PanSimulationLog = __decorate([
    customElement('pan-simulation-log')
], PanSimulationLog);
export { PanSimulationLog };
//# sourceMappingURL=pan-simulation-log.js.map