
class ButtonComponent extends HTMLElement {
  static get observedAttributes() {
    return ['color', 'background-color', 'font-size'];
  }
  constructor() {
    super();

    // Attach a shadow DOM to the element.
    const shadow = this.attachShadow({ mode: 'closed' });

    const template = document.createElement("template")

    template.innerHTML = `
      <style>
      :host{
        display: block;
        width: 100%;
      }
      ::slotted(button){
          width: 100%;
          font-family: Aruma;
          font-size: 25px;
          border-radius: 5px;
          color: white;
          padding: 10px;
          background-color: seagreen;
        }
      ::slotted(button:hover) {
          background-color: darkolivegreen;
      }
      
      </style>
        <slot name="button">
        Button
      </slot>
`


    htmx.process(template) // Tell HTMX about this component's shadow DOM
    shadow.appendChild(template.content.cloneNode(true))

    this.shadow = shadow;


  }

  connectedCallback() {
    // Apply initial attribute values
  }

  attributeChangedCallback(name, oldValue, newValue) {
  }

}


customElements.define('button-component', ButtonComponent);



