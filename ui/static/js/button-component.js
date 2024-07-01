
class ButtonComponent extends HTMLElement {
  static get observedAttributes() {
    return ['color', 'background-color', 'font-size'];
  }
  constructor() {
    super();

    // Attach a shadow DOM to the element.
    const shadow = this.attachShadow({ mode: 'open' });

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
          color: antiquewhite;
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


    shadow.appendChild(template.content.cloneNode(true))
    this.shadow = shadow;
  }

  connectedCallback() {
    // Apply initial attribute values
    htmx.process(this.shadow) // Tell HTMX about this component's shadow DOM
    this.updateStyles();
  }

  attributeChangedCallback(name, oldValue, newValue) {
    if (oldValue !== newValue) {
      this.updateStyles();
    }
  }

  updateStyles() {
    const content = this.shadowRoot.getElementById('button-element');
    const attributes = ['color', 'background-color', 'font-size']

    for (let attribute of attributes) {
      if (this.hasAttribute(attribute)) {
        const cssProperty = attribute.replace(/-(.)/g, (_, group1) => group1.toUpperCase());
        content.style[cssProperty] = this.getAttribute(attribute);
      }
    }
  }
}


customElements.define('button-component', ButtonComponent);



