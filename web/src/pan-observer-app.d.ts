import { LitElement } from 'lit';
import './components/pan-agent-card';
import './components/pan-simulation-log';
import './components/pan-trace-inspector';
import './components/pan-world-state';
export declare class PanObserverApp extends LitElement {
    static styles: import("lit").CSSResult;
    private events;
    private agents;
    private selectedTrace;
    private currentStep;
    private premise;
    private autoscroll;
    private completed;
    private paused;
    private scenarios;
    private selectedScenario;
    constructor();
    private fetchInitialState;
    private sendControl;
    private launchScenario;
    private connect;
    render(): import("lit-html").TemplateResult<1>;
}
