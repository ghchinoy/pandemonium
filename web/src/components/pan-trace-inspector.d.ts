import { LitElement } from 'lit';
import { Event } from '../types';
export declare class PanTraceInspector extends LitElement {
    static styles: import("lit").CSSResult;
    trace: Event | null;
    render(): import("lit-html").TemplateResult<1>;
    private _handleClose;
}
