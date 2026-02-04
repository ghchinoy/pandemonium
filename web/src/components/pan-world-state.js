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
let PanWorldState = class PanWorldState extends LitElement {
    constructor() {
        super(...arguments);
        this.step = 0;
        this.premise = '';
        this.completed = false;
    }
    static { this.styles = css `
    :host { display: block; background: #1a1a1a; padding: 15px; border-radius: 12px; border: 1px solid #333; margin-bottom: 20px; }
    .label { font-size: 0.7em; font-weight: bold; color: #555; text-transform: uppercase; margin-bottom: 8px; display: block; }
    .step { font-size: 1.5em; font-weight: bold; color: #ee6ff8; margin-bottom: 10px; }
    .premise { font-size: 0.85em; color: #ccc; line-height: 1.4; border-top: 1px solid #222; padding-top: 10px; }
  `; }
    render() {
        return html `
      <span class="label">Current Status</span>
      <div class="step">
        ${this.completed ? `COMPLETED (${this.step} STEPS)` : `STEP ${this.step}`}
      </div>
      <div class="premise"><strong>Premise:</strong> ${this.premise}</div>
    `;
    }
};
__decorate([
    property({ type: Number }),
    __metadata("design:type", Object)
], PanWorldState.prototype, "step", void 0);
__decorate([
    property({ type: String }),
    __metadata("design:type", Object)
], PanWorldState.prototype, "premise", void 0);
__decorate([
    property({ type: Boolean }),
    __metadata("design:type", Object)
], PanWorldState.prototype, "completed", void 0);
PanWorldState = __decorate([
    customElement('pan-world-state')
], PanWorldState);
export { PanWorldState };
//# sourceMappingURL=pan-world-state.js.map