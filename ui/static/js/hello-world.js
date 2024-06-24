
class HelloWorld extends HTMLElement {
  constructor() {
    super();

    // Attach a shadow DOM to the element.
    const shadow = this.attachShadow({ mode: 'open' });

    // Create a wrapper div
    const wrapper = document.createElement('button');

    // Apply some styles to the wrapper
    const style = document.createElement('style');
    style.textContent = `
      div {
        font-family: Arial, sans-serif;
        color: blue;
        padding: 10px;
        border: 2px solid #ccc;
        border-radius: 5px;
        display: inline-block;
        background-color: #f4f4f4;
      }
    `;

    wrapper.addEventListener('click', () => {
      alert('Hello, World!');
    });

    // Set the content of the component
    wrapper.textContent = 'Hello, World!';

    // Attach the created elements to the shadow DOM
    shadow.appendChild(style);
    shadow.appendChild(wrapper);


  }
}


customElements.define('hello-world', HelloWorld);
