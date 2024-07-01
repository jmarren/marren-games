


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
        border-right: 4px solid black;
        background-color: seagreen;
        padding-top: 100px;
      }
      nav {
        display: flex;
        flex-direction: column;
        justify-content: space-around;
      }
      .side-bar-item {
        text-align: center;
        padding-left: 20px;
        padding-right: 20px;
        padding-top: 20px;
        padding-bottom: 20px;
        border-bottom: 2px solid darkseagreen;
        font-family: Cheto;
        font-size: 30px;
        color: cornsilk;
        }
      .side-bar-item:hover {
        background-color: darkolivegreen;
      }
        
      .side-bar-item:nth-of-type(1) {
        border-top: 2px solid darkseagreen;
      }
      

      </style> 

        <div id="side-bar-container"> 
          <nav>
            <div id="profile" class="side-bar-item">Profile</div>
            <div id="games" class="side-bar-item">Games</div>
            <div id="friends" class="side-bar-item">Friends</div>
          </nav>
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



