import { LitElement, html, css } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { Event, AgentStates } from './types';

import './components/pan-agent-card';
import './components/pan-simulation-log';
import './components/pan-trace-inspector';
import './components/pan-world-state';

@customElement('pan-observer-app')
export class PanObserverApp extends LitElement {
  static styles = css`
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
  `;

  @state() private events: Event[] = [];
  @state() private agents: AgentStates = {};
  @state() private selectedTrace: Event | null = null;
  @state() private currentStep = 0;
  @state() private premise = "";
  @state() private autoscroll = true;
  @state() private completed = false;

  constructor() {
    super();
    this.connect();
  }

  private connect() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const ws = new WebSocket(protocol + '//' + window.location.host + '/ws');
    ws.onmessage = (msg) => {
      const event: Event = JSON.parse(msg.data);
      
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
        } else if (event.type === 'observation') {
          currentStatus = 'Observing';
        } else if (event.type === 'action') {
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
    return html`
      <header>
        <h1>PANDEMONIUM <span style="color: #444; font-weight: normal;">| OBSERVER</span></h1>
        <div class="status-area">
          <label class="toggle">
            <input type="checkbox" ?checked="${this.autoscroll}" @change="${(e: any) => this.autoscroll = e.target.checked}">
            <div class="slider"></div>
            Autoscroll
          </label>
          <div style="color: #04b575;">● SYSTEM LIVE</div>
        </div>
      </header>
      <main class="content ${this.selectedTrace ? 'with-inspector' : ''}" 
            @select-trace="${(e: CustomEvent) => this.selectedTrace = e.detail}"
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
            ${Object.entries(this.agents).map(([name, state]) => html`
              <pan-agent-card .name="${name}" .status="${state.status}" .isRemote="${state.isRemote}"></pan-agent-card>
            `)}
          </div>
        </aside>

        ${this.selectedTrace ? html`
          <pan-trace-inspector .trace="${this.selectedTrace}"></pan-trace-inspector>
        ` : ''}
      </main>
    `;
  }
}