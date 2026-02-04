import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';

@customElement('pan-world-state')
export class PanWorldState extends LitElement {
  static styles = css`
    :host { display: block; background: #1a1a1a; padding: 15px; border-radius: 12px; border: 1px solid #333; margin-bottom: 20px; }
    .label { font-size: 0.7em; font-weight: bold; color: #555; text-transform: uppercase; margin-bottom: 8px; display: block; }
    .step { font-size: 1.5em; font-weight: bold; color: #ee6ff8; margin-bottom: 10px; }
    .premise { font-size: 0.85em; color: #ccc; line-height: 1.4; border-top: 1px solid #222; padding-top: 10px; }
  `;

  @property({ type: Number }) step = 0;
  @property({ type: String }) premise = '';
  @property({ type: Boolean }) completed = false;

  render() {
    return html`
      <span class="label">Current Status</span>
      <div class="step">
        ${this.completed ? `COMPLETED (${this.step} STEPS)` : `STEP ${this.step}`}
      </div>
      <div class="premise"><strong>Premise:</strong> ${this.premise}</div>
    `;
  }
}
