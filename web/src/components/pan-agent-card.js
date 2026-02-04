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
let PanAgentCard = class PanAgentCard extends LitElement {
    constructor() {
        super(...arguments);
        this.name = '';
        this.status = '';
        this.isRemote = false;
    }
    static { this.styles = css `
    :host { display: block; padding: 12px; margin-bottom: 10px; background: #1a1a1a; border-radius: 8px; border: 1px solid #333; position: relative; }
    .header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 4px; }
    .name { color: #04b575; font-weight: bold; font-size: 0.9em; text-transform: uppercase; }
    .status { font-size: 0.8em; color: #888; font-style: italic; line-height: 1.4; }
    .a2a-icon { width: 16px; height: 16px; color: #7d56f4; }
  `; }
    render() {
        return html `
      <div class="header">
        <span class="name">${this.name}</span>
        ${this.isRemote ? html `
          <img class="a2a-icon" src="/a2a-logo.svg" title="Remote A2A Agent" />
        ` : ''}
      </div>
      <div class="status">${this.status}</div>
    `;
    }
};
__decorate([
    property({ type: String }),
    __metadata("design:type", Object)
], PanAgentCard.prototype, "name", void 0);
__decorate([
    property({ type: String }),
    __metadata("design:type", Object)
], PanAgentCard.prototype, "status", void 0);
__decorate([
    property({ type: Boolean }),
    __metadata("design:type", Object)
], PanAgentCard.prototype, "isRemote", void 0);
PanAgentCard = __decorate([
    customElement('pan-agent-card')
], PanAgentCard);
export { PanAgentCard };
//# sourceMappingURL=pan-agent-card.js.map