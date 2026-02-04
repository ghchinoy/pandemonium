import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { Event } from '../types';

@customElement('pan-simulation-log')
export class PanSimulationLog extends LitElement {
  static styles = css`
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
  `;

  @property({ type: Array }) events: Event[] = [];
  @property({ type: Boolean }) autoscroll = true;

  protected updated(changedProperties: Map<string, any>) {
    if (this.autoscroll && changedProperties.has('events')) {
      this.scrollTop = this.scrollHeight;
    }
  }

  render() {
    const displayEvents = this.events.filter(e => e.type !== 'step_start' && e.type !== 'sim_start');
    
    return html`
      ${displayEvents.map(e => html`
        <div class="event ${e.type}" @click="${() => this.handleEventClick(e)}">
          <span class="actor">${e.actor || 'World'}</span>
          <span class="type">(${e.type})</span>
          ${e.data ? html`<button class="inspect-btn">Inspect</button>` : ''}
          <div class="message">${e.message}</div>
        </div>
      `)}
    `;
  }

  private handleEventClick(e: Event) {
    if (e.data) {
      this.dispatchEvent(new CustomEvent('select-trace', {
        detail: e,
        bubbles: true,
        composed: true
      }));
    }
  }
}