import { LitElement } from 'lit';
import { Event } from '../types';
export declare class PanSimulationLog extends LitElement {
    static styles: import("lit").CSSResult;
    events: Event[];
    autoscroll: boolean;
    protected updated(changedProperties: Map<string, any>): void;
    render(): import("lit-html").TemplateResult<1>;
    private handleEventClick;
}
