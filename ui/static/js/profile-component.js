

class ProfileComponent extends HTMLElement {
  constructor() {
    super();

    // Attach a shadow DOM to the element.
    const shadow = this.attachShadow({ mode: 'open' });

    // Create a wrapper div
    // const wrapper = document.createElement('button');

    // Apply some styles to the wrapper
    const style = document.createElement('style');
    style.textContent = `
      :host(#shadow-root) {
        height: 100px;
        width: 100px;
        background-color: white;
        }
      ::slotted(#container) {
        height: 500px;
        padding: 50px 50px 20px;
        // width: 100%;
        border: 2px solid black;
        font-size: 40px;
        color: orange;
        font-family: "Jersey 15", sans-serif;
      }
    `;

    const template = document.getElementById("profile-template")


    const slot = document.createElement('slot')
    slot.style = style
    // Attach the created elements to the shadow DOM
    shadow.appendChild(style);

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


customElements.define('profile-component', ProfileComponent);
