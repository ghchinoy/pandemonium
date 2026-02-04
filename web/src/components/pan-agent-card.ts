import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';

@customElement('pan-agent-card')
export class PanAgentCard extends LitElement {
  static styles = css`
    :host { display: block; padding: 12px; margin-bottom: 10px; background: #1a1a1a; border-radius: 8px; border: 1px solid #333; position: relative; }
    .header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 4px; }
    .name { color: #04b575; font-weight: bold; font-size: 0.9em; text-transform: uppercase; }
    .status { font-size: 0.8em; color: #888; font-style: italic; line-height: 1.4; }
    .a2a-icon { width: 16px; height: 16px; color: #7d56f4; }
  `;

  @property({ type: String }) name = '';
  @property({ type: String }) status = '';
  @property({ type: Boolean }) isRemote = false;

  render() {
    return html`
      <div class="header">
        <span class="name">${this.name}</span>
        ${this.isRemote ? html`
          <img class="a2a-icon" src="/a2a-logo.svg" title="Remote A2A Agent" />
        ` : ''}
      </div>
      <div class="status">${this.status}</div>
    `;
    }
  }
  