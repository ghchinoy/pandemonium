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
let PanTraceInspector = class PanTraceInspector extends LitElement {
    constructor() {
        super(...arguments);
        this.trace = null;
    }
    static { this.styles = css `
    :host { display: block; background: #1a1a1a; border-left: 2px solid #7d56f4; padding: 25px; height: 100%; overflow-y: auto; box-sizing: border-box; box-shadow: -10px 0 30px rgba(0,0,0,0.5); }
    h2 { color: #7d56f4; font-size: 1.1em; margin-top: 0; display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #333; padding-bottom: 15px; }
    .close-btn { background: none; border: none; color: #555; cursor: pointer; font-size: 1.5em; }
    .close-btn:hover { color: #fff; }
    .component { margin-bottom: 25px; }
    .comp-name { font-family: 'JetBrains Mono', monospace; font-size: 0.75em; color: #04b575; text-transform: uppercase; margin-bottom: 8px; border-bottom: 1px solid #222; display: block; padding-bottom: 4px; }
    .comp-value { font-size: 0.9em; color: #aaa; line-height: 1.6; white-space: pre-wrap; }
  `; }
    render() {
        if (!this.trace)
            return html `<div></div>`;
        return html `
      <h2>
        Reasoning: ${this.trace.actor}
        <button class="close-btn" @click="${this._handleClose}">&times;</button>
      </h2>
      ${Object.entries(this.trace.data || {}).map(([name, value]) => html `
        <div class="component">
          <span class="comp-name">${name}</span>
          <div class="comp-value">${value}</div>
        </div>
      `)}
    `;
    }
    _handleClose() {
        this.dispatchEvent(new CustomEvent('close-trace', { bubbles: true, composed: true }));
    }
};
__decorate([
    property({ type: Object }),
    __metadata("design:type", Object)
], PanTraceInspector.prototype, "trace", void 0);
PanTraceInspector = __decorate([
    customElement('pan-trace-inspector')
], PanTraceInspector);
export { PanTraceInspector };
//# sourceMappingURL=pan-trace-inspector.js.map