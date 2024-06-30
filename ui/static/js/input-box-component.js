


class InputBoxComponent extends HTMLElement {
  constructor() {
    super();
    const shadow = this.attachShadow({ mode: 'open' });
    const template = document.createElement('template')

    template.innerHTML = `
      <style>
      #container {
          display: flex;
          justify-content: space-around;
          padding: 40px;
          background-color: goldenrod;
          border-radius: 10px;
          font-family: Manila;
          font-size: 40px;
          flex-direction: column;
          overflow: hidden;
          margin: 0;
      }
      ::slotted(textarea) {
          height: 40px;
          border: 3px solid black;
          border-radius: 5px;
          font-family: Manila;
          font-size: 30px;
          color: darkslategray;
          padding: 10px;
      }
      </style>
      <div id="container" >
        <div id="prompt-text" >
        What are you wondering?
        </div>
        <slot name="input-slot" ></slot>
    </div>
      `;
    shadow.appendChild(template.content.cloneNode(true))
  }
  connectedCallback() {
    console.log("Custom element added to page.");
  }

  disconnectedCallback() {
    console.log("Custom element removed from page.");
  }

  adoptedCallback() {
    console.log("Custom element moved to new page.");
  }

  attributeChangedCallback(name, oldValue, newValue) {
    console.log(`Attribute ${name} has changed.`);
  }
}


customElements.define('input-box-component', InputBoxComponent);
