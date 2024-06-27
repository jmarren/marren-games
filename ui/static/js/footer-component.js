


class FooterComponent extends HTMLElement {
  constructor() {
    super();

    // Attach a shadow DOM to the element.
    const shadow = this.attachShadow({ mode: 'open' });

    const template = document.createElement("template")

    template.innerHTML = `
      <style>
        #footer-element {
          background: green;
        }
      </style>
      <footer id="footer-element">
        This is the footer
      </footer>
`




    shadow.appendChild(template.content.cloneNode(true))

    // shadow.appendChild(wrapper);







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
