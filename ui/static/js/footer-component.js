class FooterComponent extends HTMLElement {
  constructor() {
    super();

    // Attach a shadow DOM to the element.
    const shadow = this.attachShadow({ mode: 'open' });
    const template = document.createElement("template")

    template.innerHTML = `
      <style>
        #footer-element {
          margin: 0;
          left: 0;
          bottom: 0;
          background-color: skyblue;
          width: 100%;
          height: 58.5px;
          border-top:  4px solid black;
          position: fixed;
        }
      </style>
      <div id="footer-container" >
      <footer id="footer-element">
      </footer>
      </div>
`
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


customElements.define('footer-component', FooterComponent);