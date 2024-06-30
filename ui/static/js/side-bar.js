


class SideBar extends HTMLElement {
  constructor() {
    super();

    // Attach a shadow DOM to the element.
    const shadow = this.attachShadow({ mode: 'open' });

    const template = document.createElement("template")

    template.innerHTML = `
      <style>
        #side-bar-container {
          position: fixed;
          top: 50px;
          left: 0;
          z-index: -1;
          height: 100%;
          width: 125px;
          border-right: 4px solid black;
          background-color: seagreen;
      }

      </style> 

        <div id="side-bar-container"> 
          
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


customElements.define('side-bar', SideBar);



