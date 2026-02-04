import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { Event } from '../types';

@customElement('pan-trace-inspector')
export class PanTraceInspector extends LitElement {
  static styles = css`
    :host { display: block; background: #1a1a1a; border-left: 2px solid #7d56f4; padding: 25px; height: 100%; overflow-y: auto; box-sizing: border-box; box-shadow: -10px 0 30px rgba(0,0,0,0.5); }
    h2 { color: #7d56f4; font-size: 1.1em; margin-top: 0; display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #333; padding-bottom: 15px; }
    .close-btn { background: none; border: none; color: #555; cursor: pointer; font-size: 1.5em; }
    .close-btn:hover { color: #fff; }
    .component { margin-bottom: 25px; }
    .comp-name { font-family: 'JetBrains Mono', monospace; font-size: 0.75em; color: #04b575; text-transform: uppercase; margin-bottom: 8px; border-bottom: 1px solid #222; display: block; padding-bottom: 4px; }
    .comp-value { font-size: 0.9em; color: #aaa; line-height: 1.6; white-space: pre-wrap; }
  `;

  @property({ type: Object }) trace: Event | null = null;

  render() {
    if (!this.trace) return html`<div></div>`;

    return html`
      <h2>
        Reasoning: ${this.trace.actor}
        <button class="close-btn" @click="${this._handleClose}">&times;</button>
      </h2>
      ${Object.entries(this.trace.data || {}).map(([name, value]) => html`
        <div class="component">
          <span class="comp-name">${name}</span>
          <div class="comp-value">${value}</div>
        </div>
      `)}
    `;
  }

  private _handleClose() {
    this.dispatchEvent(new CustomEvent('close-trace', { bubbles: true, composed: true }));
  }
}
