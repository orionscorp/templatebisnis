import Head from 'next/head';
import { useState, useEffect } from 'react';
import axios from 'axios';
import 'bootstrap/dist/css/bootstrap.min.css';

const HomePage = () => {
  const [paymentMethod, setPaymentMethod] = useState('bank_transfer');
  const [formData, setFormData] = useState({
    name: '',
    phone: '',
    email: '',
    bonus: false,
  });

  const basePrice = 137000;
  const bonusPrice = 57000;
  const totalPrice = basePrice + (formData.bonus ? bonusPrice : 0);

  const handleInputChange = (e) => {
    const { name, value, type, checked } = e.target;
    const finalValue = type === 'checkbox' ? checked : value;
    setFormData({ ...formData, [name]: finalValue });
  };

  const handlePaymentMethodChange = (e) => {
    setPaymentMethod(e.target.value);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      const response = await axios.post('http://localhost:8080/api/payment', { ...formData, paymentMethod, totalPrice });
      if (response.data.status === 'success') {
        alert('Detail pembayaran telah terkirim ke email');
      } else {
        alert('Gagal mengirim email pembayaran');
      }
    } catch (error) {
      console.error('Payment error:', error);
      alert('Terjadi kesalahan pada pembayaran');
    }
  };

  useEffect(() => {
    const script = document.createElement('script');
    script.src = 'https://app.sandbox.midtrans.com/snap/snap.js';
    script.setAttribute('data-client-key', 'YOUR_MIDTRANS_CLIENT_KEY');
    document.body.appendChild(script);
  }, []);

  const paymentMethods = [
    // { value: 'bank_transfer', label: 'Bank Transfer', logo: 'https://www.freepnglogos.com/uploads/bank-png/bank-building-logo-18.png' },
    { value: 'credit_card', label: 'Credit Card', logo: 'https://seeklogo.com/images/V/VISA-logo-A32D589D31-seeklogo.com.png' },
    { value: 'gopay', label: 'GoPay', logo: 'https://seeklogo.com/images/G/gopay-logo-D27C1EBD0D-seeklogo.com.png' },
    { value: 'bca_va', label: 'BCA Virtual Account', logo: 'https://seeklogo.com/images/B/bca-bank-logo-1E89320DC2-seeklogo.com.png' },
    { value: 'bni_va', label: 'BNI Virtual Account', logo: 'https://seeklogo.com/images/B/bank-bni-logo-737EE0F32C-seeklogo.com.png' },
    { value: 'echannel', label: 'Mandiri Bill (eChannel)', logo: 'https://seeklogo.com/images/B/bank_mandiri-logo-4F6233ABCC-seeklogo.com.png' },
    { value: 'cstore_indomaret', label: 'Indomaret', logo: 'https://seeklogo.com/images/I/indomaret-logo-EE717AAD0D-seeklogo.com.png' },
    { value: 'cstore_alfamart', label: 'Alfamart', logo: 'https://seeklogo.com/images/A/alfamart-logo-653AD66E16-seeklogo.com.png' },
    { value: 'qris', label: 'QRIS', logo: 'https://seeklogo.com/images/Q/quick-response-code-indonesia-standard-qris-logo-F300D5EB32-seeklogo.com.png' },
  ];

  return (
    <div className="container mt-5">
      <Head>
        <title>Payment Page</title>
      </Head>
      <h1 className="mb-4">Spreadsheet & Powerpoint Flexi DSP - 018</h1>
      <p>Sempurna untuk industri apa pun: Manajer Proyek, Pemilik usaha kecil, Pemimpin tim, Spesialis Penjaminan Mutu, HR Manager, Financial Analyst, Marketing Team, Educator & Trainers, Freelancer.</p>
      <p>Template yang menghemat waktu anda dan mempermudah segalanya.</p>

      <form onSubmit={handleSubmit} className="mt-4">
        <div className="mb-3">
          <label className="form-label">
            Nama Anda:
            <input type="text" className="form-control" name="name" value={formData.name} onChange={handleInputChange} required />
          </label>
        </div>
        <div className="mb-3">
          <label className="form-label">
            No. WhatsApp Anda:
            <input type="tel" className="form-control" name="phone" value={formData.phone} onChange={handleInputChange} required />
          </label>
        </div>
        <div className="mb-3">
          <label className="form-label">
            Email Anda:
            <input type="email" className="form-control" name="email" value={formData.email} onChange={handleInputChange} required />
          </label>
        </div>
        <div className="mb-3 form-check">
          <label className="form-check-label">
            <input type="checkbox" className="form-check-input" name="bonus" checked={formData.bonus} onChange={handleInputChange} />
            Dapatkan Bonus (+Rp 57.000)
          </label>
        </div>

        <h2>Rincian Pesanan:</h2>
        <p>Harga: Rp 137.000</p>
        <p>Bonus: {formData.bonus ? 'Ya (+Rp 57.000)' : 'Tidak'}</p>
        <p>Total: Rp {totalPrice}</p>

        <h2>Metode Pembayaran:</h2>
        {paymentMethods.map(method => (
          <div className="form-check mb-2" key={method.value}>
            <label className="form-check-label d-flex align-items-center">
              <input type="radio" className="form-check-input me-2" name="paymentMethod" value={method.value} checked={paymentMethod === method.value} onChange={handlePaymentMethodChange} />
              <img src={method.logo} alt={method.label} className="me-2" style={{ width: '40px', height: 'auto' }} />
              {method.label}
            </label>
          </div>
        ))}

        <button type="submit" className="btn btn-primary mt-3">Bayar Sekarang</button>
      </form>
    </div>
  );
};

export default HomePage;